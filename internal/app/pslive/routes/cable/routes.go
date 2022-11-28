// Package cable contains the route handlers serving Action Cables over WebSockets
// by implementing the Action Cable Protocol (https://docs.anycable.io/misc/action_cable_protocol)
// on the server.
package cable

import (
	"github.com/gorilla/websocket"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

type Handlers struct {
	r godest.TemplateRenderer

	ss *session.Store
	cc *session.CSRFTokenChecker

	acc *actioncable.Cancellers
	acs actioncable.Signer
	tsb *turbostreams.Broker
	vsb *videostreams.Broker

	wsu websocket.Upgrader

	l godest.Logger
}

func New(
	r godest.TemplateRenderer, ss *session.Store, cc *session.CSRFTokenChecker,
	acc *actioncable.Cancellers, acs actioncable.Signer, tsb *turbostreams.Broker,
	vsb *videostreams.Broker, l godest.Logger,
) *Handlers {
	return &Handlers{
		r:   r,
		ss:  ss,
		cc:  cc,
		acc: acc,
		acs: acs,
		tsb: tsb,
		vsb: vsb,
		wsu: websocket.Upgrader{
			Subprotocols: []string{actioncable.ActionCableV1MsgpackSubprotocol},
			// TODO: add parameters to the upgrader as needed
		},
		l: l,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/cable", auth.HandleHTTPWithSession(h.HandleCableGet(), h.ss))
	er.GET("/video-cable", auth.HandleHTTPWithSession(h.HandleVideoCableGet(), h.ss))
}
