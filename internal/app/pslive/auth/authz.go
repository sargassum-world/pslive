package auth

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/pkg/godest/opa"
)

// Authorization

func (a Auth) Authorized() bool {
	// Right now there's only one user who can be authenticated, namely the admin, so this is
	// good enough for now.
	return a.Identity.Authenticated
}

func (a Auth) RequireAuthz(
	ctx context.Context, input map[string]interface{}, opc *opa.Client,
) error {
	allow, remainingQueries, evalErr := opc.EvalAllow(ctx, input)
	if evalErr != nil {
		return errors.Wrap(evalErr, "couldn't check authorization")
	}
	if allow {
		return nil
	}
	if remainingQueries != nil {
		// TODO: evaluate the remaining queries
		return errors.New("authorization depends on remaining queries, which are not yet implemented")
	}

	authzErr, evalErr := opc.EvalErrors(ctx, input)
	if evalErr != nil {
		return errors.Wrap(evalErr, "couldn't check authorization errors")
	}
	if authzErr == nil {
		return errors.New("unauthorized but missing error message")
	}
	return authzErr
}

// HTTP

func (a Auth) RequireHTTPAuthz(c echo.Context, opc *opa.Client) error {
	return a.RequireAuthz(
		c.Request().Context(),
		opa.NewRouteInput(
			c.Request().Method, c.Request().URL.RequestURI(), a.Identity.User, a.Identity.Authenticated,
		),
		opc,
	)
}

func RequireHTTPAuthz(ss session.Store, opc *opa.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			a, _, err := GetFromRequest(c.Request(), ss, c.Logger())
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s",
					c.Request().Method, c.Request().URL.RequestURI(),
				)
			}
			if err = a.RequireHTTPAuthz(c, opc); err != nil {
				// We return StatusNotFound instead of StatusUnauthorized or StatusForbidden to avoid
				// leaking information about the existence of secret resources.
				return echo.NewHTTPError(http.StatusNotFound, errors.Wrap(
					err, "couldn't authorize on http route",
				))
			}
			return next(c)
		}
	}
}

// Turbo Streams

func (a Auth) RequireTSAuthz(c turbostreams.Context, opc *opa.Client) error {
	if c.Method() == turbostreams.MethodUnsub || c.Method() == turbostreams.MethodPub {
		// We can't prevent unsubscription; and closing a tab triggers an unsubscription while also
		// canceling context, which will interrupt policy evaluation (and cause an evalErr).
		// So unsubscription is always authorized.
		// The server is always authorized to handle pub.
		return nil
	}

	return a.RequireAuthz(
		c.Context(),
		opa.NewRouteInput(
			c.Method(), c.Topic(), a.Identity.User, a.Identity.Authenticated,
		),
		opc,
	)
}

func RequireTSAuthz(ss session.Store, opc *opa.Client) turbostreams.MiddlewareFunc {
	return func(next turbostreams.HandlerFunc) turbostreams.HandlerFunc {
		return func(c turbostreams.Context) error {
			a, _, err := LookupStored(c.SessionID(), ss)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s",
					c.Method(), c.Topic(),
				)
			}
			if err = a.RequireTSAuthz(c, opc); err != nil {
				return errors.Wrapf(err, "couldn't authorize %s on %s", c.Method(), c.Topic())
			}
			return next(c)
		}
	}
}
