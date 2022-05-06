// Package users contains the route handlers related to users.
package users

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc *ory.Client

	tsh *turbostreams.MessagesHub

	ps *presence.Store
	cs *chat.Store
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, tsh *turbostreams.MessagesHub,
	ps *presence.Store, cs *chat.Store,
) *Handlers {
	return &Handlers{
		r:   r,
		oc:  oc,
		tsh: tsh,
		ps:  ps,
		cs:  cs,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	hr.GET("/users", h.HandleUsersGet())
	hr.GET("/users/:id", h.HandleUserGet())
	// TODO: make and use a middleware which checks to ensure the user exists
	tsr.SUB("/users/:id/chat/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps))
	tsr.UNSUB("/users/:id/chat/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.MSG("/users/:id/chat/users", handling.HandleTSMsg(h.r, ss))
	tsr.SUB("/users/:id/chat/messages", turbostreams.EmptyHandler)
	tsr.MSG("/users/:id/chat/messages", handling.HandleTSMsg(h.r, ss))
	// TODO: add a paginated GET handler for chat messages to support chat history infiniscroll
	// TODO: make and use a middleware which checks to ensure the user exists
	hr.POST("/users/:id/chat/messages", handling.HandleChatMessagesPost(h.r, h.oc, h.tsh, h.cs), haz)
}
