package instruments

import (
	"fmt"
	"net/http"

	"github.com/atrox/haikunatorgo"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentsGet() auth.HTTPHandlerFunc {
	t := "instruments/instruments.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		instruments, err := h.is.GetInstruments(c.Request().Context())
		if err != nil {
			return err
		}
		// TODO: we should adapt it into a []InstrumentViewData or something

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, instruments, a)
	}
}

func (h *Handlers) HandleInstrumentsPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		i := instruments.Instrument{
			Name:        haikunator.New().Haikunate(),
			Description: "An unknown instrument!",
			AdminID:     a.Identity.User,
		}
		id, err := h.is.AddInstrument(c.Request().Context(), i)
		if err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}
