package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

// Authorization

func (a Auth) Authorized() bool {
	// Right now there's only one user who can be authenticated, namely the admin, so this is
	// good enough for now.
	return a.Identity.Authenticated
}

// HTTP

func (a Auth) RequireHTTPAuthz() error {
	if a.Authorized() {
		return nil
	}

	// We return StatusNotFound instead of StatusUnauthorized or StatusForbidden to avoid leaking
	// information about the existence of secret resources.
	// TODO: would the error message leak information? If so, we should leave it blank everywhere
	// across the app.
	return echo.NewHTTPError(http.StatusNotFound, "unauthorized")
}

func RequireHTTPAuthz(ss session.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			a, _, err := GetWithSession(c.Request(), ss, c.Logger())
			if err != nil {
				return err
			}
			if err = a.RequireHTTPAuthz(); err != nil {
				return err
			}
			return next(c)
		}
	}
}

// Turbo Streams

func (a Auth) RequireTSAuthz() error {
	if a.Authorized() {
		return nil
	}

	if a.Identity.User == "" {
		return errors.New("unknown user not authorized")
	}
	return errors.Errorf("user %s not authorized", a.Identity.User)
}

func RequireTSAuthz(ss session.Store) turbostreams.MiddlewareFunc {
	return func(next turbostreams.HandlerFunc) turbostreams.HandlerFunc {
		return func(c turbostreams.Context) error {
			sess, err := ss.Lookup(c.SessionID())
			if err != nil {
				return errors.Errorf("couldn't lookup session to check authz on %s", c.Topic())
			}
			if sess == nil {
				return errors.Errorf("unknown user not authorized on %s", c.Topic())
			}
			a, err := GetWithoutRequest(*sess, ss)
			if err != nil {
				return errors.Wrap(err, "couldn't lookup auth info for session")
			}
			if err = a.RequireTSAuthz(); err != nil {
				return errors.Wrapf(err, "couldn't authorize on %s", c.Topic())
			}
			return next(c)
		}
	}
}
