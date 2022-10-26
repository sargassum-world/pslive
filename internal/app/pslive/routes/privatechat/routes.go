// Package privatechat contains the route handlers related to private chats.
package privatechat

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc  *ory.Client
	azc *auth.AuthzChecker

	tsh *turbostreams.Hub

	ps *presence.Store
	cs *chat.Store
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, azc *auth.AuthzChecker, tsh *turbostreams.Hub,
	ps *presence.Store, cs *chat.Store,
) *Handlers {
	return &Handlers{
		r:   r,
		oc:  oc,
		azc: azc,
		tsh: tsh,
		ps:  ps,
		cs:  cs,
	}
}

func (h *Handlers) Register(
	er godest.EchoRouter, tsr turbostreams.Router, ss *session.Store,
) {
	hr := auth.NewHTTPRouter(er, ss)
	// TODO: make and use a middleware which checks to ensure the users exist
	tsr.SUB(
		"/private-chats/:first/:second/chat/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps),
	)
	tsr.UNSUB("/private-chats/:first/:second/chat/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.MSG("/private-chats/:first/:second/chat/users", handling.HandleTSMsg(h.r, ss))
	tsr.SUB("/private-chats/:first/:second/chat/messages", turbostreams.EmptyHandler)
	tsr.MSG("/private-chats/:first/:second/chat/messages", handling.HandleTSMsg(h.r, ss))
	// TODO: add a paginated GET handler for chat messages to support chat history infiniscroll
	// TODO: make the paginated GET handler check for user authorization to view the chat history
	hr.POST("/private-chats/:first/:second/chat/messages", handling.HandleChatMessagesPost(
		h.r, h.oc, h.azc, h.tsh, h.cs,
	))
}
