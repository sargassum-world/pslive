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

	is  *instruments.Store
	pco *planktoscope.Orchestrator
	ps  *presence.Store
	cs  *chat.Store
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, tsh *turbostreams.MessagesHub,
	is *instruments.Store, pco *planktoscope.Orchestrator, ps *presence.Store, cs *chat.Store,
) *Handlers {
	return &Handlers{
		r:   r,
		oc:  oc,
		tsh: tsh,
		is:  is,
		pco: pco,
		ps:  ps,
		cs:  cs,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	hr.GET("/instruments", h.HandleInstrumentsGet())
	hr.POST("/instruments", h.HandleInstrumentsPost(), haz)
	hr.GET("/instruments/:id", h.HandleInstrumentGet())
	hr.POST("/instruments/:id", h.HandleInstrumentPost(), haz)
	hr.POST("/instruments/:id/name", h.HandleInstrumentNamePost(), haz)
	hr.POST("/instruments/:id/description", h.HandleInstrumentDescriptionPost(), haz)
	// TODO: make and use a middleware which checks to ensure the instrument exists
	tsr.SUB("/instruments/:id/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps))
	tsr.UNSUB("/instruments/:id/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.MSG("/instruments/:id/users", handling.HandleTSMsg(h.r, ss))
	hr.POST("/instruments/:id/cameras", h.HandleInstrumentCamerasPost(), haz)
	hr.POST("/instruments/:id/cameras/:cameraID", h.HandleInstrumentCameraPost(), haz)
	hr.POST("/instruments/:id/controllers", h.HandleInstrumentControllersPost(), haz)
	hr.POST("/instruments/:id/controllers/:controllerID", h.HandleInstrumentControllerPost(), haz)
	tsr.SUB("/instruments/:id/controllers/:controllerID/pump", turbostreams.EmptyHandler)
	tsr.PUB("/instruments/:id/controllers/:controllerID/pump", h.HandlePumpPub())
	tsr.MSG("/instruments/:id/controllers/:controllerID/pump", handling.HandleTSMsg(h.r, ss))
	hr.POST("/instruments/:id/controllers/:controllerID/pump", h.HandlePumpPost())
	// TODO: make and use a middleware which checks to ensure the instrument exists
	tsr.SUB("/instruments/:id/chat/messages", turbostreams.EmptyHandler)
	tsr.MSG("/instruments/:id/chat/messages", handling.HandleTSMsg(h.r, ss))
	// TODO: add a paginated GET handler for chat messages to support chat history infiniscroll
	// TODO: make and use a middleware which checks to ensure the instrument exists
	hr.POST("/instruments/:id/chat/messages", handling.HandleChatMessagesPost(
		h.r, h.oc, h.tsh, h.cs,
	), haz)
}
