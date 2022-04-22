// Package instruments contains the route handlers related to imaging instruments.
package instruments

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type Handlers struct {
	r godest.TemplateRenderer

	tsh *turbostreams.MessagesHub

	ic  *instruments.Client
	pcs map[string]*planktoscope.Client
	oc  *ory.Client
}

func New(
	r godest.TemplateRenderer, tsh *turbostreams.MessagesHub,
	ic *instruments.Client, pcs map[string]*planktoscope.Client, oc *ory.Client,
) *Handlers {
	return &Handlers{
		r:   r,
		tsh: tsh,
		ic:  ic,
		pcs: pcs,
		oc:  oc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	hr.GET("/instruments", h.HandleInstrumentsGet())
	hr.GET("/instruments/:name", h.HandleInstrumentGet())
	tsr.SUB("/instruments/:name/controller/pump", turbostreams.EmptyHandler)
	tsr.PUB("/instruments/:name/controller/pump", h.HandlePumpPub())
	tsr.MSG("/instruments/:name/controller/pump", handling.HandleTSMsg(h.r, ss))
	hr.POST("/instruments/:name/controller/pump", h.HandlePumpPost(), haz)
}
