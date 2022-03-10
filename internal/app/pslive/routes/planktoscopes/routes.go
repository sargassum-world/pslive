// Package planktoscopes contains the route handlers related to Planktoscopes.
package planktoscopes

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/planktoscopes"
)

type Handlers struct {
	r  godest.TemplateRenderer
	pc *planktoscopes.Client
}

func New(r godest.TemplateRenderer, pc *planktoscopes.Client) *Handlers {
	return &Handlers{
		r:  r,
		pc: pc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, sc *session.Client) {
	ar := auth.NewRouter(er, sc)
	ar.GET("/planktoscopes", h.HandlePlanktoscopesGet())
	ar.GET("/planktoscopes/:name", h.HandlePlanktoscopeGet())
}
