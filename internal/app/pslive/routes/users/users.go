package users

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

func (h *Handlers) HandleUsersGet() auth.HTTPHandlerFunc {
	t := "users/users.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		var identities []ory.Identity
		if !h.oc.Config.NoAuth {
			oryIdentities, err := h.oc.GetIdentities(c.Request().Context())
			if err != nil {
				return err
			}
			identities = oryIdentities
		}
		if !h.ac.Config.NoAuth || (h.oc.Config.NoAuth && h.ac.Config.NoAuth) {
			identities = append(identities, ory.Identity{
				ID:         ory.IdentityID(h.ac.Config.AdminUsername),
				Identifier: ory.IdentityIdentifier(h.ac.Config.AdminUsername),
				Email:      "none",
			})
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, identities, a)
	}
}
