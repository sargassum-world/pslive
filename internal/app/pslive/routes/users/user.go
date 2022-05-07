package users

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type UserViewData struct {
	Identity                ory.Identity
	PublicKnownViewers      []presence.User
	PublicAnonymousViewers  []string
	PublicChatMessages      []handling.ChatMessageViewData
	PrivateKnownViewers     []presence.User
	PrivateAnonymousViewers []string
	PrivateChatMessages     []handling.ChatMessageViewData
}

func getUserViewData(
	ctx context.Context, id string, a auth.Auth, oc *ory.Client, ps *presence.Store, cs *chat.Store,
) (*UserViewData, error) {
	identity, err := oc.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}

	// Public chat
	publicKnown, publicAnonymous := ps.List("/users/" + id + "/chat/users")
	publicMessages, err := cs.GetMessagesByTopic(
		ctx, "/users/"+id+"/chat/messages", chat.DefaultMessagesLimit,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get public chat messages for user %s", id)
	}
	publicMessagesAdapted, err := handling.AdaptChatMessages(ctx, publicMessages, oc)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't adapt public chat messages for user %s into view data", id,
		)
	}

	// Private chat
	var privateKnown []presence.User
	var privateAnon []string
	var privateMessages []chat.Message
	var privateMessagesAdapted []handling.ChatMessageViewData
	if a.Identity.Authenticated && a.Identity.User != id {
		first := id
		second := a.Identity.User
		if second < first {
			first, second = second, first
		}
		privateKnown, privateAnon = ps.List("/private-chats/" + first + "/" + second + "/chat/users")
		if privateMessages, err = cs.GetMessagesByTopic(
			ctx, "/private-chats/"+first+"/"+second+"/chat/messages", chat.DefaultMessagesLimit,
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't get private chat messages for users %s & %s", first, second,
			)
		}
		if privateMessagesAdapted, err = handling.AdaptChatMessages(
			ctx, privateMessages, oc,
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't adapt private chat messages for user %s into view data", id,
			)
		}
	}

	return &UserViewData{
		Identity:                identity,
		PublicKnownViewers:      publicKnown,
		PublicAnonymousViewers:  publicAnonymous,
		PublicChatMessages:      publicMessagesAdapted,
		PrivateKnownViewers:     privateKnown,
		PrivateAnonymousViewers: privateAnon,
		PrivateChatMessages:     privateMessagesAdapted,
	}, nil
}

func (h *Handlers) HandleUserGet() auth.HTTPHandlerFunc {
	t := "users/user.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")

		// Run queries
		userViewData, err := getUserViewData(c.Request().Context(), id, a, h.oc, h.ps, h.cs)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *userViewData, a)
	}
}
