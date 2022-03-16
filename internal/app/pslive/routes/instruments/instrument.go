package instruments

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type InstrumentData struct {
	Instrument instruments.Instrument
	Controller planktoscope.Planktoscope
}

func getInstrumentData(
	name string, ic *instruments.Client, pcs map[string]*planktoscope.Client,
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

	pc, ok := pcs[instrument.Controller]
	if !ok {
		return nil, errors.Errorf("planktoscope client for instrument %s not found", name)
	}
	return &InstrumentData{
		Instrument: *instrument,
		Controller: pc.GetState(),
	}, nil
}

func (h *Handlers) HandleInstrumentGet() auth.Handler {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		instrumentData, err := getInstrumentData(name, h.ic, h.pcs)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *instrumentData, a)
	}
}

// Pumping

func (h *Handlers) HandleInstrumentPumpPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Negotiate response content type
		fmt.Println(c.Request().Header["Accept"])

		// Parse params
		name := c.Param("name")
		pumping := strings.ToLower(c.FormValue("pumping")) == "start"

		// Run queries
		instrument, err := h.ic.FindInstrument(name)
		if err != nil {
			return err
		}
		pc, ok := h.pcs[instrument.Controller]
		if !ok {
			return errors.Errorf("planktoscope client for instrument %s not found", name)
		}
		var token mqtt.Token
		if !pumping {
			if token, err = pc.StopPump(); err != nil {
				return err
			}
		} else {
			// TODO: use echo's request binding functionality instead of strconv.ParseFloat
			forward := strings.ToLower(c.FormValue("direction")) == "forward"
			const floatWidth = 64
			volume, err := strconv.ParseFloat(c.FormValue("volume"), floatWidth)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse volume"))
			}
			flowrate, err := strconv.ParseFloat(c.FormValue("flowrate"), floatWidth)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse flowrate"))
			}
			if token, err = pc.StartPump(forward, volume, flowrate); err != nil {
				return err
			}
		}

		stateUpdated := pc.PumpStateBroadcasted()
		// TODO: instead of waiting forever, have a timeout before redirecting and displaying a
		// warning message that we haven't heard any pump state updates from the planktoscope
		if token.Wait(); token.Error() != nil {
			return token.Error()
		}
		<-stateUpdated

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%s", name))
	}
}
