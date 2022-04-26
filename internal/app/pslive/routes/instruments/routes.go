// Package instruments contains the route handlers related to imaging instruments.
package instruments

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc *ory.Client

	tsh *turbostreams.MessagesHub

	ic  *instruments.Client
	pcs map[string]*planktoscope.Client
	ps  *presence.Store
	cs  *chat.Store
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, tsh *turbostreams.MessagesHub,
	ic *instruments.Client, pcs map[string]*planktoscope.Client, ps *presence.Store, cs *chat.Store,
) *Handlers {
	return &Handlers{
		r:   r,
		oc:  oc,
		tsh: tsh,
		ic:  ic,
		pcs: pcs,
		ps:  ps,
		cs:  cs,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	hr.GET("/instruments", h.HandleInstrumentsGet())
	hr.GET("/instruments/:name", h.HandleInstrumentGet())
	// TODO: make and use a middleware which checks to ensure the instrument exists
	tsr.SUB("/instruments/:name/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps))
	tsr.UNSUB("/instruments/:name/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.MSG("/instruments/:name/users", handling.HandleTSMsg(h.r, ss))
	tsr.SUB("/instruments/:name/controller/pump", turbostreams.EmptyHandler)
	tsr.PUB("/instruments/:name/controller/pump", h.HandlePumpPub())
	tsr.MSG("/instruments/:name/controller/pump", handling.HandleTSMsg(h.r, ss))
	hr.POST("/instruments/:name/controller/pump", h.HandlePumpPost(), haz)
	// TODO: make and use a middleware which checks to ensure the instrument exists
	tsr.SUB("/instruments/:name/chat/messages", turbostreams.EmptyHandler)
	tsr.MSG("/instruments/:name/chat/messages", handling.HandleTSMsg(h.r, ss))
	// TODO: make and use a middleware which checks to ensure the instrument exists
	hr.POST("/instruments/:name/chat/messages", handling.HandleChatMessagesPost(
		h.r, h.oc, h.tsh, h.cs,
	), haz)
}
