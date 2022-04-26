package users

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type UserData struct {
	Identity ory.Identity
}

func getUserData(ctx context.Context, id string, oc *ory.Client) (*UserData, error) {
	identity, err := oc.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserData{
		Identity: identity,
	}, nil
}

func (h *Handlers) HandleUserGet() auth.HTTPHandlerFunc {
	t := "users/user.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")

		// Run queries
		userData, err := getUserData(c.Request().Context(), id, h.oc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *userData, a)
	}
}
