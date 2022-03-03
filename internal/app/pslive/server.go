// Package pslive provides the Planktoscope Live server.
package pslive

import (
	"fmt"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/unrolled/secure"

	"github.com/sargassum-world/pslive/internal/app/pslive/client"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/assets"
	"github.com/sargassum-world/pslive/internal/app/pslive/tmplfunc"
	imw "github.com/sargassum-world/pslive/internal/middleware"
	"github.com/sargassum-world/fluitans/pkg/godest"
	gmw "github.com/sargassum-world/fluitans/pkg/godest/middleware"
	"github.com/sargassum-world/pslive/web"
)

type Server struct {
	Embeds   godest.Embeds
	Inlines  godest.Inlines
	Renderer godest.TemplateRenderer
	Globals  *client.Globals
	Handlers *routes.Handlers
}

func NewServer(e *echo.Echo) (s *Server, err error) {
	s = &Server{}
	s.Embeds = web.NewEmbeds()
	s.Inlines = web.NewInlines()
	s.Renderer, err = godest.NewTemplateRenderer(
		s.Embeds, s.Inlines, sprig.FuncMap(), tmplfunc.FuncMap(
			tmplfunc.NewHashedNamers(assets.AppURLPrefix, assets.StaticURLPrefix, s.Embeds),
		),
	)
	if err != nil {
		s = nil
		err = errors.Wrap(err, "couldn't make template renderer")
		return
	}

	s.Globals, err = client.NewGlobals(e.Logger)
	if err != nil {
		s = nil
		err = errors.Wrap(err, "couldn't make app globals")
		return
	}

	s.Handlers = routes.New(s.Renderer, s.Globals.Clients)
	return
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
	e.Use(s.Globals.Clients.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(s.Renderer, e.Logger, s.Globals.Clients.Sessions)),
	))
	e.Use(imw.RequireContentTypes(echo.MIMEApplicationForm))
	// TODO: enable Prometheus and rate-limiting

	// Handlers
	e.HTTPErrorHandler = NewHTTPErrorHandler(s.Renderer, s.Globals.Clients.Sessions)
	s.Handlers.Register(e, s.Embeds)
}
