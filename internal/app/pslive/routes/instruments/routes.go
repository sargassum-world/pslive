// Package instruments contains the route handlers related to imaging instruments.
package instruments

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type Handlers struct {
	r   godest.TemplateRenderer
	ic  *instruments.Client
	pcs map[string]*planktoscope.Client
}

func New(
	r godest.TemplateRenderer, ic *instruments.Client, pcs map[string]*planktoscope.Client,
) *Handlers {
	return &Handlers{
		r:   r,
		ic:  ic,
		pcs: pcs,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, sc *session.Client) {
	ar := auth.NewRouter(er, sc)
	az := auth.RequireAuthz(sc)
	ar.GET("/instruments", h.HandleInstrumentsGet())
	ar.GET("/instruments/:name", h.HandleInstrumentGet())
	ar.POST("/instruments/:name/pump", h.HandleInstrumentPumpPost(), az)
}
