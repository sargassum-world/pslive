package instruments

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type InstrumentData struct {
	Instrument instruments.Instrument
	Controller planktoscope.Planktoscope
}

func getInstrumentData(
	name string, ic *instruments.Client, pc *planktoscope.Client,
) (*InstrumentData, error) {
	instrument, err := ic.FindInstrument(name)
	if err != nil {
		return nil, err
	}
	if instrument == nil {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %s not found", name),
		)
	}

	// TODO: select the planktoscope client based on the instrument
	planktoscope := pc.GetState()

	return &InstrumentData{
		Instrument: *instrument,
		Controller: planktoscope,
	}, nil
}

func (h *Handlers) HandleInstrumentGet() auth.Handler {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		instrumentData, err := getInstrumentData(name, h.ic, h.pc)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks before computing etag for client-side caching
		return h.r.CacheablePage(c.Response(), c.Request(), t, *instrumentData, a)
	}
}
