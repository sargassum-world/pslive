// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Handlers struct {
	r godest.TemplateRenderer

	ss *session.Store
	oc *ory.Client

	acc *actioncable.Cancellers

	ps *presence.Store

	l godest.Logger
}

func New(
	r godest.TemplateRenderer, ss *session.Store, oc *ory.Client, acc *actioncable.Cancellers,
	ps *presence.Store, l godest.Logger,
) *Handlers {
	return &Handlers{
		r:   r,
		ss:  ss,
		oc:  oc,
		acc: acc,
		ps:  ps,
		l:   l,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/csrf", h.HandleCSRFGet())
	er.GET("/login", auth.HandleHTTPWithSession(h.HandleLoginGet(), h.ss))
	er.POST("/sessions", h.HandleSessionsPost())
}
