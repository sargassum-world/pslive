// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

type Handlers struct {
	r  godest.TemplateRenderer
	ac *authn.Client
	sc *session.Client
}

func New(r godest.TemplateRenderer, ac *authn.Client, sc *session.Client) *Handlers {
	return &Handlers{
		r:  r,
		ac: ac,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/csrf", h.HandleCSRFGet())
	er.GET("/login", auth.HandleWithSession(h.HandleLoginGet(), h.sc))
	er.POST("/sessions", h.HandleSessionsPost())
}
