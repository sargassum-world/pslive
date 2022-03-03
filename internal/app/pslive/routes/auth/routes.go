// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-world/fluitans/pkg/godest"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/authn"
	"github.com/sargassum-world/pslive/internal/clients/sessions"
)

type Handlers struct {
	r  godest.TemplateRenderer
	ac *authn.Client
	sc *sessions.Client
}

func New(r godest.TemplateRenderer, ac *authn.Client, sc *sessions.Client) *Handlers {
	return &Handlers{
		r:  r,
		ac: ac,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	ar := auth.NewAuthAwareRouter(er, h.sc)
	er.GET("/csrf", h.HandleCSRFGet())
	ar.GET("/login", h.HandleLoginGet())
	er.POST("/sessions", h.HandleSessionsPost())
}
