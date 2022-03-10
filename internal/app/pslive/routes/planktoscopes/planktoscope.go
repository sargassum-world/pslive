package planktoscopes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/planktoscopes"
)

type PlanktoscopeData struct {
	Planktoscope planktoscopes.Planktoscope
}

func getPlanktoscopeData(name string, pc *planktoscopes.Client) (*PlanktoscopeData, error) {
	planktoscope, err := pc.FindPlanktoscope(name)
	if err != nil {
		return nil, err
	}
	if planktoscope == nil {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("planktoscope %s not found", name),
		)
	}

	return &PlanktoscopeData{
		Planktoscope: *planktoscope,
	}, nil
}

func (h *Handlers) HandlePlanktoscopeGet() auth.Handler {
	t := "planktoscopes/planktoscope.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		planktoscopeData, err := getPlanktoscopeData(name, h.pc)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks before computing etag for client-side caching
		return h.r.CacheablePage(c.Response(), c.Request(), t, *planktoscopeData, a)
	}
}
