package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/open-policy-agent/opa/ast"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"
	"zombiezen.com/go/sqlite"

	"github.com/sargassum-world/pslive/pkg/godest/database"
	"github.com/sargassum-world/pslive/pkg/godest/opa"
)

// Authorization

type AuthzChecker struct {
	db  *database.DB
	opc *opa.Client
	t   opa.SQLiteTranspiler
}

func NewAuthzChecker(db *database.DB, opc *opa.Client) *AuthzChecker {
	return &AuthzChecker{
		db:  db,
		opc: opc,
		t:   opa.NewSQLiteTranspiler("input.context.db"),
	}
}

func (azc *AuthzChecker) requireAuthzWithoutContextualData(
	ctx context.Context, input map[string]interface{},
) (authzErr error, evalErr error) {
	// Policy-reported errors are only for policy/route matching errors and other SQL-independent
	// errors; they are not for reporting authz denial reasons, because error generation doesn't
	// work well with partial evaluation
	reportedErr, evalErr := azc.opc.EvalError(ctx, input)
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
	allow, evalErr := azc.opc.EvalAllow(ctx, input)
	if errors.Is(evalErr, context.Canceled) {
		return nil, evalErr
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for authz")
	}
	if !allow {
		return errors.New("not authorized by policy without contextual data"), nil
	}
	return nil, nil // authorized!
}

func (azc *AuthzChecker) evaluateRemainingQueries(
	ctx context.Context, remaining []ast.Body,
) (result bool, err error) {
	statement, err := azc.t.Parse(remaining)
	if err != nil {
		return false, errors.Wrap(err, "couldn't translate remaining queries into SQL")
	}
	if err := azc.db.ExecuteSelection(
		ctx, statement.String(), statement.NamedParams(), func(s *sqlite.Stmt) error {
			result = s.GetBool(statement.ResultName)
			return nil
		},
	); err != nil {
		return false, errors.Wrap(err, "Couldn't execute SQL selection")
	}
	return result, nil
}

func (azc *AuthzChecker) requireAuthzWithContextualData(
	ctx context.Context, input map[string]interface{},
) (authzErr error, evalErr error) {
	allow, remainingQueries, evalErr := azc.opc.EvalAllowContextual(ctx, input)
	if errors.Is(evalErr, context.Canceled) {
		return nil, evalErr
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for authz")
	}
	if allow {
		return nil, nil // authorized!
	}
	if remainingQueries == nil {
		return errors.New("unauthorized according to policy"), nil
	}

	result, err := azc.evaluateRemainingQueries(ctx, remainingQueries)
	if err != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate remaining rego queries")
	}
	if !result {
		return errors.New("unauthorized according to policy with contextual data"), nil
	}
	return nil, nil
}

func (azc *AuthzChecker) RequireAuthz(
	ctx context.Context, input map[string]interface{},
) (authzErr error, evalErr error) {
	authzErr, evalErr = azc.requireAuthzWithoutContextualData(ctx, input)
	if evalErr != nil {
		return nil, evalErr
	}
	if authzErr == nil { // allowed by policy even without contextual data
		return nil, nil
	}

	// This is much slower, but we do it if we must
	return azc.requireAuthzWithContextualData(ctx, input)
}

func (azc *AuthzChecker) Allow(
	ctx context.Context, a Auth, resourcePath, operationMethod string, operationParams interface{},
) (bool, error) {
	authzErr, evalErr := azc.RequireAuthz(ctx, opa.Input{
		Resource:  opa.NewResource(resourcePath),
		Operation: opa.NewOperation(operationMethod, operationParams),
		Subject:   a.Identity.NewSubject(),
	}.Map())
	if errors.Is(evalErr, context.Canceled) {
		return false, evalErr
	}
	if evalErr != nil {
		return false, errors.Wrapf(evalErr, "couldn't check authz")
	}
	return authzErr == nil, nil
}

func (azc *AuthzChecker) RequireHTTPAuthz(c echo.Context, a Auth) (authzErr error, evalErr error) {
	formParams, err := c.FormParams()
	if err != nil {
		return nil, errors.New("couldn't parse form params for input to authz check")
	}
	return azc.RequireAuthz(
		c.Request().Context(),
		opa.Input{
			Resource:  opa.NewResource(c.Request().URL.RequestURI()),
			Operation: opa.NewOperation(c.Request().Method, formParams),
			Subject:   a.Identity.NewSubject(),
		}.Map(),
	)
}

func (azc *AuthzChecker) RequireTSAuthz(
	c turbostreams.Context, a Auth,
) (authzErr error, evalErr error) {
	if c.Method() == turbostreams.MethodUnsub || c.Method() == turbostreams.MethodPub {
		// We can't prevent unsubscription; and closing a tab triggers an unsubscription while also
		// canceling context, which will interrupt policy evaluation (and cause an evalErr).
		// So unsubscription is always authorized.
		// The server is always authorized to handle pub.
		return nil, nil
	}

	return azc.RequireAuthz(
		c.Context(),
		opa.Input{
			Resource:  opa.NewResource(c.Topic()),
			Operation: opa.NewOperation(c.Method(), url.Values{}),
			Subject:   a.Identity.NewSubject(),
		}.Map(),
	)
}

func (azc *AuthzChecker) NewHTTPMiddleware(ss session.Store) echo.MiddlewareFunc {
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
			authzErr, evalErr := azc.RequireHTTPAuthz(c, a)
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

func (azc *AuthzChecker) NewTSMiddleware(ss session.Store) turbostreams.MiddlewareFunc {
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
			authzErr, evalErr := azc.RequireTSAuthz(c, a)
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
