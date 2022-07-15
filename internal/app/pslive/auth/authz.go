package auth

import (
	"context"
	"net/http"
	"net/url"

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
) (authzErr error, checkErr error) {
	// Policy-reported errors are only for policy/route matching errors and other SQL-independent
	// errors; they are not for reporting authz denial reasons, because error generation doesn't
	// work well with partial evaluation
	reportedErr, evalErr := opc.EvalError(ctx, input)
	if errors.Is(evalErr, context.Canceled) {
		return nil, evalErr
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for error messages")
	}
	if reportedErr != nil {
		return nil, errors.Wrap(reportedErr, "policy reported error")
	}

	// This is a performance optimization for allowed routes which don't depend on a SQL lookup
	// (e.g. for static assets)
	allow, evalErr := opc.EvalAllow(ctx, input)
	if errors.Is(evalErr, context.Canceled) {
		return nil, evalErr
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for authz")
	}
	if allow {
		return nil, nil // authorized!
	}

	// This is much slower, but we do it if we must
	allow, remainingQueries, evalErr := opc.EvalAllowContextual(ctx, input)
	if errors.Is(evalErr, context.Canceled) {
		return nil, evalErr
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for authz")
	}
	if allow {
		return nil, nil // authorized!
	}
	if remainingQueries != nil {
		// TODO: evaluate the remaining queries
		return errors.New("authz depends on remaining queries, which are not yet implemented"), nil
	}

	return errors.New("unauthorized according to policy"), nil
}

// HTTP

func (a Auth) RequireHTTPAuthz(c echo.Context, opc *opa.Client) (authzErr error, checkErr error) {
	formParams, err := c.FormParams()
	if err != nil {
		return nil, errors.New("couldn't parse form params for input to authz check")
	}
	return a.RequireAuthz(
		c.Request().Context(),
		opa.Input{
			Resource:  opa.NewResource(c.Request().URL.RequestURI()),
			Operation: opa.NewOperation(c.Request().Method, formParams),
			Subject:   a.Identity.NewSubject(),
		}.Map(),
		opc,
	)
}

func RequireHTTPAuthz(ss session.Store, opc *opa.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			uri := c.Request().URL.RequestURI()

			a, _, err := GetFromRequest(c.Request(), ss, c.Logger())
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s", method, uri,
				)
			}
			if _, err := c.FormParams(); err != nil {
				// We check this before RequireHTTPAuthz so we can return HTTP 400
				return echo.NewHTTPError(http.StatusBadRequest, errors.Errorf(
					"couldn't parse form params for input to authz check on %s %s", method, uri,
				))
			}
			authzErr, evalErr := a.RequireHTTPAuthz(c, opc)
			if errors.Is(evalErr, context.Canceled) {
				return evalErr
			}
			if evalErr != nil {
				return errors.Wrapf(evalErr, "couldn't check authz on %s %s", method, uri)
			}
			if authzErr != nil {
				// We return StatusNotFound instead of StatusUnauthorized or StatusForbidden to avoid
				// leaking information about the existence of secret resources.
				return echo.NewHTTPError(http.StatusNotFound, errors.Wrapf(
					authzErr, "couldn't authorize %s %s", method, uri,
				))
			}
			return next(c)
		}
	}
}

// Turbo Streams

func (a Auth) RequireTSAuthz(
	c turbostreams.Context, opc *opa.Client,
) (authzErr error, checkErr error) {
	if c.Method() == turbostreams.MethodUnsub || c.Method() == turbostreams.MethodPub {
		// We can't prevent unsubscription; and closing a tab triggers an unsubscription while also
		// canceling context, which will interrupt policy evaluation (and cause an evalErr).
		// So unsubscription is always authorized.
		// The server is always authorized to handle pub.
		return nil, nil
	}

	return a.RequireAuthz(
		c.Context(),
		opa.Input{
			Resource:  opa.NewResource(c.Topic()),
			Operation: opa.NewOperation(c.Method(), url.Values{}),
			Subject:   a.Identity.NewSubject(),
		}.Map(),
		opc,
	)
}

func RequireTSAuthz(ss session.Store, opc *opa.Client) turbostreams.MiddlewareFunc {
	return func(next turbostreams.HandlerFunc) turbostreams.HandlerFunc {
		return func(c turbostreams.Context) error {
			method := c.Method()
			topic := c.Topic()

			a, _, err := LookupStored(c.SessionID(), ss)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't lookup auth info for session to check authz on %s %s", method, topic,
				)
			}
			authzErr, evalErr := a.RequireTSAuthz(c, opc)
			if errors.Is(evalErr, context.Canceled) {
				return evalErr
			}
			if evalErr != nil {
				return errors.Wrapf(evalErr, "couldn't check authz on %s %s", method, topic)
			}
			if authzErr != nil {
				return errors.Wrapf(authzErr, "couldn't authorize %s %s", method, topic)
			}
			return next(c)
		}
	}
}
