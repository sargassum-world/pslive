package instruments

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

func (h *Handlers) HandleInstrumentsGet() auth.HTTPHandlerFunc {
	t := "instruments/instruments.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		instruments, err := h.ic.GetInstruments()
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, instruments, a)
	}
}
