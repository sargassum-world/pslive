// Package pslive provides the Planktoscope Live server.
package pslive

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	gmw "github.com/sargassum-world/fluitans/pkg/godest/middleware"
	"github.com/unrolled/secure"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/pslive/db"
	"github.com/sargassum-world/pslive/internal/app/pslive/client"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/assets"
	"github.com/sargassum-world/pslive/internal/app/pslive/tmplfunc"
	"github.com/sargassum-world/pslive/internal/app/pslive/workers"
	"github.com/sargassum-world/pslive/internal/clients/database"
	"github.com/sargassum-world/pslive/web"
)

type Server struct {
	DBEmbeds database.Embeds
	Globals  *client.Globals
	Embeds   godest.Embeds
	Inlines  godest.Inlines
	Renderer godest.TemplateRenderer
	Handlers *routes.Handlers
}

func (s *Server) openDB(ctx context.Context) error {
	schema, err := s.DBEmbeds.NewSchema()
	if err != nil {
		return errors.Wrap(err, "couldn't load database schema")
	}
	if err = s.Globals.DB.Open(); err != nil {
		return errors.Wrap(err, "couldn't open connection pool for database")
	}
	// TODO: close the store when the context is canceled, in order to allow flushing the WAL
	if err = s.Globals.DB.Migrate(ctx, schema); err != nil {
		// TODO: close the store if the migration failed
		return errors.Wrap(err, "couldn't perform database schema migrations")
	}
	return nil
}

func NewServer(e *echo.Echo) (s *Server, err error) {
	s = &Server{}
	s.DBEmbeds = db.NewEmbeds()
	s.Globals, err = client.NewGlobals(s.DBEmbeds, e.Logger)
	if err != nil {
		s = nil
		return nil, errors.Wrap(err, "couldn't make app globals")
	}

	s.Embeds = web.NewEmbeds()
	s.Inlines = web.NewInlines()
	s.Renderer, err = godest.NewTemplateRenderer(
		s.Embeds, s.Inlines, sprig.FuncMap(), tmplfunc.FuncMap(
			tmplfunc.NewHashedNamers(assets.AppURLPrefix, assets.StaticURLPrefix, s.Embeds),
			s.Globals.TSSigner.Sign,
		),
	)
	if err != nil {
		s = nil
		return nil, errors.Wrap(err, "couldn't make template renderer")
	}

	s.Handlers = routes.New(s.Renderer, s.Globals)

	// TODO: opening the DB should happen when we start the server, not when we instantiate it!
	// Ideally, we'd move e.Start(...) into a background worker (and rename it to RunWorkers), on
	// the same level as TSBroker.Serve. Then we just call openDB when we run s.Start(...), which
	// opens the database connections and launches all the workers.
	err = errors.Wrap(s.openDB(context.TODO()), "couldn't open database")
	return s, err
}

func (s *Server) Register(e *echo.Echo) {
	// HTTP Headers Middleware
	csp := strings.Join([]string{
		"default-src 'self'",
		// Warning: script-src 'self' may not be safe to use if we're hosting user-uploaded content.
		// Then we'll need to provide hashes for scripts & styles we include by URL, and we'll need to
		// add the SRI integrity attribute to the tags including those files; however, it's unclear
		// how well-supported they are by browsers.
		fmt.Sprintf(
			"script-src 'self' 'unsafe-inline' %s", strings.Join(s.Inlines.ComputeJSHashesForCSP(), " "),
		),
		fmt.Sprintf(
			"style-src 'self' 'unsafe-inline' %s", strings.Join(append(
				s.Inlines.ComputeCSSHashesForCSP(),
				// Note: Turbo Drive tries to install a style tag for its progress bar, which leads to a CSP
				// error. We add a hash for it here, assuming ProgressBar.animationDuration == 300:
				"'sha512-rVca7GmrbBAUUoTnu9V9a6ZR4WAZdxFUnrsg3B+1zEsES4K6q7EW02LIXdYmE5aofGOwLySKKtOafC0hq892BA=='",
			), " "),
		),
		"object-src 'none'",
		"child-src 'self'",
		"img-src *",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		// TODO: add HTTPS-related settings for CSP, including upgrade-insecure-requests
	}, "; ")
	e.Use(echo.WrapMiddleware(secure.New(secure.Options{
		// TODO: add HTTPS options
		FrameDeny:               true,
		ContentTypeNosniff:      true,
		ContentSecurityPolicy:   csp,
		ReferrerPolicy:          "no-referrer",
		CrossOriginOpenerPolicy: "same-origin",
	}).Handler))
	e.Use(echo.WrapMiddleware(gmw.SetCORP("same-site")))
	e.Use(echo.WrapMiddleware(gmw.SetCOEP("require-corp")))

	// Compression Middleware
	e.Use(middleware.Decompress())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: s.Globals.Config.HTTP.GzipLevel,
	}))

	// Other Middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echo.WrapMiddleware(s.Globals.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(s.Renderer, e.Logger, s.Globals.Sessions)),
	)))
	e.Use(gmw.RequireContentTypes(echo.MIMEApplicationForm))
	// TODO: enable Prometheus and rate-limiting

	// Handlers
	e.HTTPErrorHandler = NewHTTPErrorHandler(s.Renderer, s.Globals.Sessions)
	s.Handlers.Register(e, s.Globals.TSBroker, s.Embeds)
}

func (s *Server) RunBackgroundWorkers(ctx context.Context) {
	eg, _ := errgroup.WithContext(ctx) // Workers run independently, so we don't need egctx
	eg.Go(func() error {
		return workers.EstablishPlanktoscopeControllerConnections(ctx, s.Globals.Planktoscopes)
	})
	eg.Go(func() error {
		return s.Globals.TSBroker.Serve(ctx)
	})
	if err := eg.Wait(); err != nil {
		s.Globals.Logger.Error(err)
	}
}
