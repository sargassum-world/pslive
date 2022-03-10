package planktoscopes

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

func (h *Handlers) HandlePlanktoscopesGet() auth.Handler {
	t := "planktoscopes/planktoscopes.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		planktoscopes, err := h.pc.GetPlanktoscopes()
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, planktoscopes, a)
	}
}
