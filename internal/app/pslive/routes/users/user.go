package users

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type UserViewData struct {
	Identity ory.Identity

	PublicKnownViewers      []presence.User
	PublicAnonymousViewers  []string
	PublicChatMessages      []handling.ChatMessageViewData
	PrivateKnownViewers     []presence.User
	PrivateAnonymousViewers []string
	PrivateChatMessages     []handling.ChatMessageViewData

	Instruments []instruments.Instrument
}

func getUserViewData(
	ctx context.Context, id string, a auth.Auth, oc *ory.Client,
	is *instruments.Store, ps *presence.Store, cs *chat.Store,
) (vd UserViewData, err error) {
	if vd.Identity, err = oc.GetIdentity(ctx, id); err != nil {
		return UserViewData{}, err
	}

	// Public chat
	vd.PublicKnownViewers, vd.PublicAnonymousViewers = ps.List("/users/" + id + "/chat/users")
	publicMessages, err := cs.GetMessagesByTopic(
		ctx, "/users/"+id+"/chat/messages", chat.DefaultMessagesLimit,
	)
	if err != nil {
		return UserViewData{}, errors.Wrapf(err, "couldn't get public chat messages for user %s", id)
	}
	if vd.PublicChatMessages, err = handling.AdaptChatMessages(ctx, publicMessages, oc); err != nil {
		return UserViewData{}, errors.Wrapf(
			err, "couldn't adapt public chat messages for user %s into view data", id,
		)
	}

	// Private chat
	if a.Identity.Authenticated && a.Identity.User != id {
		first := id
		second := a.Identity.User
		if second < first {
			first, second = second, first
		}
		vd.PrivateKnownViewers, vd.PrivateAnonymousViewers = ps.List(
			"/private-chats/" + first + "/" + second + "/chat/users",
		)
		var privateMessages []chat.Message
		privateMessages, err = cs.GetMessagesByTopic(
			ctx, "/private-chats/"+first+"/"+second+"/chat/messages", chat.DefaultMessagesLimit,
		)
		if err != nil {
			return UserViewData{}, errors.Wrapf(
				err, "couldn't get private chat messages for users %s & %s", first, second,
			)
		}
		if vd.PrivateChatMessages, err = handling.AdaptChatMessages(
			ctx, privateMessages, oc,
		); err != nil {
			return UserViewData{}, errors.Wrapf(
				err, "couldn't adapt private chat messages for user %s into view data", id,
			)
		}
	}

	// Instruments
	if vd.Instruments, err = is.GetInstrumentsByAdminID(ctx, id); err != nil {
		return UserViewData{}, err
	}
	// TODO: we should adapt it into a []InstrumentViewData or something

	return vd, nil
}

func (h *Handlers) HandleUserGet() auth.HTTPHandlerFunc {
	t := "users/user.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")

		// Run queries
		userViewData, err := getUserViewData(c.Request().Context(), id, a, h.oc, h.is, h.ps, h.cs)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, userViewData, a)
	}
}
