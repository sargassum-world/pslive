package cable

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

func (h *Handlers) HandleCableGet() auth.HTTPHandlerFuncWithSession {
	return func(c echo.Context, _ auth.Auth, sess *sessions.Session) error {
		wsc, err := h.wsu.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		const wsMaxMessageSize = 512
		wsc.SetReadLimit(wsMaxMessageSize)

		acc := actioncable.Upgrade(wsc, actioncable.WithChannels(
			map[string]actioncable.ChannelFactory{
				turbostreams.ChannelName: turbostreams.NewChannelFactory(h.tsb, sess.ID, h.tss.Check),
			},
			make(map[string]actioncable.Channel),
			actioncable.WithCSRFTokenChecker(func(token string) error {
				return h.cc.Check(c.Request(), token)
			}),
		))
		ctx, cancel := context.WithCancel(c.Request().Context())
		h.acc.Add(sess.ID, cancel)
		serr := acc.Serve(ctx)
		// We can't return errors after the HTTP request is upgraded to a websocket, so we just log them
		if serr != nil && serr != context.Canceled {
			h.l.Error(serr)
		}
		if err := acc.Close(serr); err != nil {
			h.l.Error(err)
		}
		return nil
	}
}
