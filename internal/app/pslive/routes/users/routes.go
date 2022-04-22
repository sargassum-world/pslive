// Package users contains the route handlers related to users.
package users

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc *ory.Client
}

func New(r godest.TemplateRenderer, oc *ory.Client) *Handlers {
	return &Handlers{
		r:  r,
		oc: oc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	hr.GET("/users", h.HandleUsersGet())
	hr.GET("/users/:id", h.HandleUserGet())
}
