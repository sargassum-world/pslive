package users

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

func (h *Handlers) HandleUsersGet() auth.HTTPHandlerFunc {
	t := "users/users.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		identities, err := h.oc.GetIdentities(c.Request().Context())
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, identities, a)
	}
}
