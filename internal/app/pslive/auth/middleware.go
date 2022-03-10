package auth

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type (
	Handler            func(c echo.Context, a Auth) error
	HandlerWithSession func(c echo.Context, a Auth, sess *sessions.Session) error
)

func Handle(h Handler, sc *session.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		a, sess, err := GetWithSession(c.Request(), sc, c.Logger())
		// We don't expect the handler to write to the session, so we save it now
		if serr := sess.Save(c.Request(), c.Response()); serr != nil {
			return errors.Wrap(err, "couldn't save new session to replace invalid session")
		}
		if err != nil {
			return err
		}
		return h(c, a)
	}
}

func HandleWithSession(h HandlerWithSession, sc *session.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		a, sess, err := GetWithSession(c.Request(), sc, c.Logger())
		if err != nil {
			return err
		}
		return h(c, a, sess)
	}
}

// Router is a routing adapter between echo.Handler and this package's Handler, by
// automatically extracting auth data from the session of the request.
type Router struct {
	er godest.EchoRouter
	sc *session.Client
}

func NewRouter(er godest.EchoRouter, sc *session.Client) Router {
	return Router{
		er: er,
		sc: sc,
	}
}

func (r *Router) CONNECT(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.CONNECT(path, Handle(h, r.sc), m...)
}

func (r *Router) DELETE(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.DELETE(path, Handle(h, r.sc), m...)
}

func (r *Router) GET(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.GET(path, Handle(h, r.sc), m...)
}

func (r *Router) HEAD(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.HEAD(path, Handle(h, r.sc), m...)
}

func (r *Router) OPTIONS(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.OPTIONS(path, Handle(h, r.sc), m...)
}

func (r *Router) PATCH(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.PATCH(path, Handle(h, r.sc), m...)
}

func (r *Router) POST(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.POST(path, Handle(h, r.sc), m...)
}

func (r *Router) PUT(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.PUT(path, Handle(h, r.sc), m...)
}

func (r *Router) TRACE(path string, h Handler, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.TRACE(path, Handle(h, r.sc), m...)
}
