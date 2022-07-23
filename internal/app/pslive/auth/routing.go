package auth

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

// HTTP

type (
	HTTPHandlerFunc            func(c echo.Context, a Auth) error
	HTTPHandlerFuncWithSession func(c echo.Context, a Auth, sess *sessions.Session) error
)

func HandleHTTP(h HTTPHandlerFunc, ss session.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		a, sess, err := GetFromRequest(c.Request(), ss, c.Logger())
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

func HandleHTTPWithSession(h HTTPHandlerFuncWithSession, ss session.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		a, sess, err := GetFromRequest(c.Request(), ss, c.Logger())
		if err != nil {
			return err
		}
		return h(c, a, sess)
	}
}

// HTTPRouter is a routing adapter between echo.HandlerFunc and this package's HTTPHandlerFunc, by
// automatically extracting auth data from the session of the request.
type HTTPRouter struct {
	er godest.EchoRouter
	ss session.Store
}

func NewHTTPRouter(er godest.EchoRouter, ss session.Store) HTTPRouter {
	return HTTPRouter{
		er: er,
		ss: ss,
	}
}

func (r *HTTPRouter) CONNECT(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.CONNECT(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) DELETE(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.DELETE(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) GET(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.GET(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) HEAD(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.HEAD(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) OPTIONS(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.OPTIONS(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) PATCH(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.PATCH(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) POST(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.POST(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) PUT(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.PUT(path, HandleHTTP(h, r.ss), m...)
}

func (r *HTTPRouter) TRACE(path string, h HTTPHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return r.er.TRACE(path, HandleHTTP(h, r.ss), m...)
}

// Turbo Streams

type (
	TSHandlerFunc            func(c turbostreams.Context, a Auth) error
	TSHandlerFuncWithSession func(c turbostreams.Context, a Auth, sess *sessions.Session) error
)

func HandleTS(h TSHandlerFunc, ss session.Store) turbostreams.HandlerFunc {
	return func(c turbostreams.Context) error {
		a, _, err := LookupStored(c.SessionID(), ss)
		if err != nil {
			return errors.Wrapf(
				err, "couldn't lookup auth info for session to check authz on %s", c.Topic(),
			)
		}
		return h(c, a)
	}
}

func HandleTSWithSession(h TSHandlerFuncWithSession, ss session.Store) turbostreams.HandlerFunc {
	return func(c turbostreams.Context) error {
		a, sess, err := LookupStored(c.SessionID(), ss)
		if err != nil {
			return errors.Wrapf(
				err, "couldn't lookup auth info for session to check authz on %s", c.Topic(),
			)
		}
		return h(c, a, sess)
	}
}

// TSRouter is a routing adapter between turbostreams.HandlerFunc and this package's TSHandlerFunc,
// by automatically extracting auth data from the session associated with the Action Cable
// connection underlying the Turbo Stream.
type TSRouter struct {
	tsr turbostreams.Router
	ss  session.Store
}

func NewTSRouter(tsr turbostreams.Router, ss session.Store) TSRouter {
	return TSRouter{
		tsr: tsr,
		ss:  ss,
	}
}

func (r *TSRouter) SUB(
	topic string, h TSHandlerFunc, m ...turbostreams.MiddlewareFunc,
) *turbostreams.Route {
	return r.tsr.SUB(topic, HandleTS(h, r.ss), m...)
}

func (r *TSRouter) PUB(
	topic string, h TSHandlerFunc, m ...turbostreams.MiddlewareFunc,
) *turbostreams.Route {
	return r.tsr.PUB(topic, HandleTS(h, r.ss), m...)
}

func (r *TSRouter) MSG(
	topic string, h TSHandlerFunc, m ...turbostreams.MiddlewareFunc,
) *turbostreams.Route {
	return r.tsr.MSG(topic, HandleTS(h, r.ss), m...)
}
