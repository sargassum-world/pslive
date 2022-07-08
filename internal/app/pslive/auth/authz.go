package auth

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

// Authorization

type RouteChecker func(
	ctx context.Context, method, route string, authenticated bool, identity string,
) (allow bool, authzErr error, evalErr error)

func (a Auth) Authorized() bool {
	// Right now there's only one user who can be authenticated, namely the admin, so this is
	// good enough for now.
	return a.Identity.Authenticated
}

// HTTP

func (a Auth) RequireHTTPAuthz(c echo.Context, checker RouteChecker) error {
	allow, authzErr, evalErr := checker(
		c.Request().Context(), c.Request().Method, c.Request().URL.RequestURI(),
		a.Identity.Authenticated, a.Identity.User,
	)
	if evalErr != nil {
		return errors.Wrap(evalErr, "couldn't check http route authorization")
	}
	if allow {
		return nil
	}

	// We return StatusNotFound instead of StatusUnauthorized or StatusForbidden to avoid leaking
	// information about the existence of secret resources.
	return echo.NewHTTPError(http.StatusNotFound, authzErr)
}

func RequireHTTPAuthz(ss session.Store, checker RouteChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			a, _, err := GetFromRequest(c.Request(), ss, c.Logger())
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s",
					c.Request().Method, c.Request().URL.RequestURI(),
				)
			}
			if err = a.RequireHTTPAuthz(c, checker); err != nil {
				return err // Return the raw error, which is an echo HTTPError, without wrapping it
			}
			return next(c)
		}
	}
}

// Turbo Streams

func (a Auth) RequireTSAuthz(c turbostreams.Context, checker RouteChecker) error {
	if c.Method() == turbostreams.MethodUnsub || c.Method() == turbostreams.MethodPub {
		// We can't prevent unsubscription; and closing a tab triggers an unsubscription while also
		// canceling context, which will interrupt policy evaluation (and cause an evalErr).
		// So unsubscription is always authorized.
		// The server is always authorized to handle pub.
		return nil
	}

	allow, authzErr, evalErr := checker(
		c.Context(), c.Method(), c.Topic(), a.Identity.Authenticated, a.Identity.User,
	)
	if evalErr != nil {
		return errors.Wrap(evalErr, "couldn't check turbo streams route authorization")
	}
	if allow {
		return nil
	}

	return authzErr
}

func RequireTSAuthz(ss session.Store, checker RouteChecker) turbostreams.MiddlewareFunc {
	return func(next turbostreams.HandlerFunc) turbostreams.HandlerFunc {
		return func(c turbostreams.Context) error {
			a, _, err := LookupStored(c.SessionID(), ss)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s",
					c.Method(), c.Topic(),
				)
			}
			if err = a.RequireTSAuthz(c, checker); err != nil {
				return errors.Wrapf(err, "couldn't authorize %s on %s", c.Method(), c.Topic())
			}
			return next(c)
		}
	}
}
