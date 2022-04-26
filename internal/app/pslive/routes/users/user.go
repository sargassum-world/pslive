package users

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type UserData struct {
	Identity               ory.Identity
	PublicKnownViewers     []presence.User
	PublicAnonymousViewers []string
	PublicChatMessages     []chat.Message
}

func getUserData(
	ctx context.Context, id string, oc *ory.Client, ps *presence.Store, cs *chat.Store,
) (*UserData, error) {
	identity, err := oc.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	publicKnown, publicAnonymous := ps.List("/users/" + id + "/chat/public/users")
	publicMessages := cs.List("/users/" + id + "/chat/public/messages")
	return &UserData{
		Identity:               identity,
		PublicKnownViewers:     publicKnown,
		PublicAnonymousViewers: publicAnonymous,
		PublicChatMessages:     publicMessages,
	}, nil
}

func (h *Handlers) HandleUserGet() auth.HTTPHandlerFunc {
	t := "users/user.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")

		// Run queries
		userData, err := getUserData(c.Request().Context(), id, h.oc, h.ps, h.cs)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *userData, a)
	}
}
