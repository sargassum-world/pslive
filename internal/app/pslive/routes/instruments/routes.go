// Package instruments contains the route handlers related to imaging instruments.
package instruments

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc  *ory.Client
	azc *auth.AuthzChecker

	tsh *turbostreams.Hub

	is  *instruments.Store
	pco *planktoscope.Orchestrator
	ps  *presence.Store
	cs  *chat.Store
	vsb *videostreams.Broker
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, azc *auth.AuthzChecker, tsh *turbostreams.Hub,
	is *instruments.Store, pco *planktoscope.Orchestrator,
	ps *presence.Store, cs *chat.Store, vsb *videostreams.Broker,
) *Handlers {
	return &Handlers{
		r:   r,
		oc:  oc,
		azc: azc,
		tsh: tsh,
		is:  is,
		pco: pco,
		ps:  ps,
		cs:  cs,
		vsb: vsb,
	}
}

func (h *Handlers) Register(
	er godest.EchoRouter, tsr turbostreams.Router, ss *session.Store,
) {
	hr := auth.NewHTTPRouter(er, ss)
	hr.GET("/instruments", h.HandleInstrumentsGet())
	hr.POST("/instruments", h.HandleInstrumentsPost())
	hr.GET("/instruments/:id", h.HandleInstrumentGet())
	hr.POST("/instruments/:id", h.HandleInstrumentPost())
	hr.POST("/instruments/:id/name", h.HandleInstrumentNamePost())
	hr.POST("/instruments/:id/description", h.HandleInstrumentDescriptionPost())
	tsr.SUB("/instruments/:id/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps))
	tsr.UNSUB("/instruments/:id/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.SUB("/instruments/:id/users/list", turbostreams.EmptyHandler)
	tsr.MSG("/instruments/:id/users/list", handling.HandleTSMsg(h.r, ss))
	tsr.SUB("/instruments/:id/users/count", turbostreams.EmptyHandler)
	tsr.MSG("/instruments/:id/users/count", handling.HandleTSMsg(h.r, ss))
	hr.POST("/instruments/:id/cameras", h.HandleInstrumentCamerasPost())
	hr.POST("/instruments/:id/cameras/:cameraID", h.HandleInstrumentCameraPost())
	er.GET("/instruments/:id/cameras/:cameraID/frame.jpeg", h.HandleInstrumentCameraFrameGet())
	er.GET("/instruments/:id/cameras/:cameraID/stream.mjpeg", h.HandleInstrumentCameraStreamGet())
	hr.POST("/instruments/:id/controllers", h.HandleInstrumentControllersPost())
	hr.POST("/instruments/:id/controllers/:controllerID", h.HandleInstrumentControllerPost())
	tsr.SUB("/instruments/:id/controllers/:controllerID/pump", turbostreams.EmptyHandler)
	tsr.PUB("/instruments/:id/controllers/:controllerID/pump", h.HandlePumpPub())
	tsr.MSG("/instruments/:id/controllers/:controllerID/pump", handling.HandleTSMsg(
		h.r, ss, h.ModifyPumpMsgData(),
	))
	hr.POST("/instruments/:id/controllers/:controllerID/pump", h.HandlePumpPost())
	tsr.SUB("/instruments/:id/chat/messages", turbostreams.EmptyHandler)
	tsr.MSG("/instruments/:id/chat/messages", handling.HandleTSMsg(h.r, ss))
	// TODO: add a paginated GET handler for chat messages to support chat history infiniscroll
	hr.POST("/instruments/:id/chat/messages", handling.HandleChatMessagesPost(
		h.r, h.oc, h.azc, h.tsh, h.cs,
	))
}
