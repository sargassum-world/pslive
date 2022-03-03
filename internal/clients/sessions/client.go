// Package sessions provides a high-level client for session management
package sessions

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/quasoft/memstore"

	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Client struct {
	Config Config
	Logger godest.Logger
	// TODO: allow configuration to use sqlite for a persistent session store
	Store sessions.Store
}

func (sc *Client) Get(c echo.Context) (*sessions.Session, error) {
	return session.Get(c.Request(), sc.Config.CookieName, sc.Store)
}

func (sc *Client) Regenerate(c echo.Context) (*sessions.Session, error) {
	return session.Regenerate(c.Request(), sc.Config.CookieName, sc.Store)
}

func (sc *Client) Invalidate(c echo.Context) (*sessions.Session, error) {
	return session.Invalidate(c.Request(), sc.Config.CookieName, sc.Store)
}

func (sc *Client) NewCSRFMiddleware(opts ...csrf.Option) echo.MiddlewareFunc {
	sameSite := csrf.SameSiteDefaultMode
	switch sc.Config.CookieOptions.SameSite {
	case http.SameSiteLaxMode:
		sameSite = csrf.SameSiteLaxMode
	case http.SameSiteStrictMode:
		sameSite = csrf.SameSiteStrictMode
	case http.SameSiteNoneMode:
		sameSite = csrf.SameSiteNoneMode
	}
	options := []csrf.Option{
		csrf.Path(sc.Config.CookieOptions.Path),
		csrf.Domain(sc.Config.CookieOptions.Domain),
		csrf.MaxAge(sc.Config.CookieOptions.MaxAge),
		csrf.Secure(sc.Config.CookieOptions.Secure),
		csrf.HttpOnly(sc.Config.CookieOptions.HttpOnly),
		csrf.SameSite(sameSite),
		csrf.RequestHeader(sc.Config.CSRFOptions.HeaderName),
		csrf.FieldName(sc.Config.CSRFOptions.FieldName),
	}
	options = append(options, opts...)
	return echo.WrapMiddleware(csrf.Protect(sc.Config.AuthKey, options...))
}

func NewMemStoreClient(l godest.Logger) (*Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions client config")
	}

	store := memstore.NewMemStore(config.AuthKey)
	store.Options = &config.CookieOptions

	return &Client{
		Config: config,
		Logger: l,
		Store:  store,
	}, nil
}
