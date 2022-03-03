// Package auth provides application-level standardization for authentication
package auth

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
)

type Identity struct {
	Authenticated bool
	User          string
}

type CSRFBehavior struct {
	InlineToken bool
}

type CSRF struct {
	Config   sessions.CSRFOptions
	Behavior CSRFBehavior
	Token    string
}

type Auth struct {
	Identity Identity
	CSRF     CSRF
}

// Middleware & Routing Adapter

type AuthAwareHandler func(c echo.Context, a Auth) error

func HandleWithAuth(h AuthAwareHandler, sc *sessions.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		a, _, err := GetWithSession(c, sc)
		if err != nil {
			return err
		}
		return h(c, a)
	}
}

// AuthAwareRouter is a routing adapter between echo.Handler and AuthAwareHandler, by automatically
// extracting auth data from the session of the request.
type AuthAwareRouter struct {
	er godest.EchoRouter
	sc *sessions.Client
}

func NewAuthAwareRouter(er godest.EchoRouter, sc *sessions.Client) AuthAwareRouter {
	return AuthAwareRouter{
		er: er,
		sc: sc,
	}
}

func (r *AuthAwareRouter) CONNECT(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.CONNECT(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) DELETE(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.DELETE(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) GET(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.GET(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) HEAD(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.HEAD(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) OPTIONS(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.OPTIONS(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) PATCH(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.PATCH(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) POST(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.POST(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) PUT(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.PUT(path, HandleWithAuth(h, r.sc), m...)
}

func (r *AuthAwareRouter) TRACE(
	path string, h AuthAwareHandler, m ...echo.MiddlewareFunc,
) *echo.Route {
	return r.er.TRACE(path, HandleWithAuth(h, r.sc), m...)
}
