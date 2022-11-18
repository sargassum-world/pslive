package cable

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/handling"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

func serveWSConn(
	r *http.Request, wsc *websocket.Conn, sess *sessions.Session,
	channelFactories map[string]actioncable.ChannelFactory,
	cc *session.CSRFTokenChecker, acc *actioncable.Cancellers, l godest.Logger,
) {
	conn := actioncable.Upgrade(wsc, actioncable.WithChannels(
		channelFactories, make(map[string]actioncable.Channel),
		actioncable.WithCSRFTokenChecker(func(token string) error {
			return cc.Check(r, token)
		}),
	))
	ctx, cancel := context.WithCancel(r.Context())
	acc.Add(sess.ID, cancel)
	serr := handling.Except(conn.Serve(ctx), context.Canceled)
	// We can't return errors after the HTTP request is upgraded to a websocket, so we just log them
	if serr != nil {
		l.Error(serr)
	}
	if err := conn.Close(serr); err != nil {
		l.Error(err)
	}
}

func (h *Handlers) HandleCableGet() auth.HTTPHandlerFuncWithSession {
	return func(c echo.Context, _ auth.Auth, sess *sessions.Session) error {
		wsc, err := h.wsu.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return errors.Wrap(err, "couldn't upgrade http request to websocket connection")
		}

		const wsMaxMessageSize = 512
		wsc.SetReadLimit(wsMaxMessageSize)
		serveWSConn(
			c.Request(), wsc, sess,
			map[string]actioncable.ChannelFactory{
				turbostreams.ChannelName: turbostreams.NewChannelFactory(h.tsb, sess.ID, h.tss.Check),
			},
			h.cc, h.acc, h.l,
		)
		return nil
	}
}

func (h *Handlers) HandleVideoCableGet() auth.HTTPHandlerFuncWithSession {
	return func(c echo.Context, _ auth.Auth, sess *sessions.Session) error {
		wsc, err := h.wsu.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return errors.Wrap(err, "couldn't upgrade http request to websocket connection")
		}

		const wsMaxMessageSize = 512
		wsc.SetReadLimit(wsMaxMessageSize)
		// TODO: make this action cable connection use msgpack instead of json for efficiency reasons
		serveWSConn(
			c.Request(), wsc, sess,
			map[string]actioncable.ChannelFactory{
				videostreams.ChannelName: videostreams.NewChannelFactory(h.vsb, sess.ID, h.l, h.tss.Check),
			},
			h.cc, h.acc, h.l,
		)
		return nil
	}
}
