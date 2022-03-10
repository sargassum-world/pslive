package instruments

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

type InstrumentData struct {
	Instrument instruments.Instrument
}

func getInstrumentData(name string, pc *instruments.Client) (*InstrumentData, error) {
	instrument, err := pc.FindInstrument(name)
	if err != nil {
		return nil, err
	}
	if instrument == nil {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %s not found", name),
		)
	}

	return &InstrumentData{
		Instrument: *instrument,
	}, nil
}

func (h *Handlers) HandleInstrumentGet() auth.Handler {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		instrumentData, err := getInstrumentData(name, h.pc)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks before computing etag for client-side caching
		return h.r.CacheablePage(c.Response(), c.Request(), t, *instrumentData, a)
	}
}
