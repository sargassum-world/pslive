// Package privatechat contains the route handlers related to private chats.
package privatechat

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
	tsaz := auth.RequireTSAuthz(ss)
	// TODO: make and use a middleware which checks to ensure the users exist
	tsr.SUB(
		"/private-chats/:first/:second/chat/users", handling.HandlePresenceSub(h.r, ss, h.oc, h.ps),
		tsaz, // FIXME: currently any authenticated user can subscribe!
	)
	tsr.UNSUB("/private-chats/:first/:second/chat/users", handling.HandlePresenceUnsub(h.r, ss, h.ps))
	tsr.MSG("/private-chats/:first/:second/chat/users", handling.HandleTSMsg(h.r, ss))
	tsr.SUB("/private-chats/:first/:second/chat/messages", turbostreams.EmptyHandler, tsaz)
	tsr.MSG("/private-chats/:first/:second/chat/messages", handling.HandleTSMsg(h.r, ss))
	hr.POST("/private-chats/:first/:second/chat/messages", handling.HandleChatMessagesPost(
		h.r, h.oc, h.tsh, h.cs,
	), haz) // FIXME: currently any authenticated user can send a message!
}
