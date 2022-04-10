package instruments

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

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

func (h *Handlers) HandleInstrumentGet() auth.HTTPHandlerFunc {
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

func handlePumpSettings(
	pumpingRaw, direction, volumeRaw, flowrateRaw string, pc *planktoscope.Client,
) (err error) {
	pumping := (strings.ToLower(pumpingRaw) == "start") || (strings.ToLower(pumpingRaw) == "restart")
	var token mqtt.Token
	if !pumping {
		if token, err = pc.StopPump(); err != nil {
			return err
		}
	} else {
		// TODO: use echo's request binding functionality instead of strconv.ParseFloat
		// TODO: perform input validation and handle invalid inputs
		forward := strings.ToLower(direction) == "forward"
		const floatWidth = 64
		volume, err := strconv.ParseFloat(volumeRaw, floatWidth)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse volume"))
		}
		flowrate, err := strconv.ParseFloat(flowrateRaw, floatWidth)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse flowrate"))
		}
		if token, err = pc.StartPump(forward, volume, flowrate); err != nil {
			return err
		}
	}

	stateUpdated := pc.PumpStateBroadcasted()
	// TODO: instead of waiting forever, have a timeout before redirecting and displaying a
	// warning message that we haven't heard any pump state updates from the planktoscope.
	if token.Wait(); token.Error() != nil {
		return token.Error()
	}
	<-stateUpdated
	return nil
}

func (h *Handlers) HandleInstrumentPumpPost() auth.HTTPHandlerFunc {
	t := "instruments/planktoscope/pump.partial.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		instrument, err := h.ic.FindInstrument(name)
		if err != nil {
			return err
		}
		pc, ok := h.pcs[instrument.Controller]
		if !ok {
			return errors.Errorf("planktoscope client for instrument %s not found", name)
		}
		if err = handlePumpSettings(
			c.FormValue("pumping"), c.FormValue("direction"),
			c.FormValue("volume"), c.FormValue("flowrate"), pc,
		); err != nil {
			return err
		}

		state := pc.GetState()
		// TODO: also/instead broadcast when the mqtt broker pushes out a state update
		message := turbostreams.Message{
			Action:   turbostreams.ActionReplace,
			Target:   "/instruments/" + name + "/controller/pump",
			Template: t,
			Data: map[string]interface{}{
				"Instrument":   instrument,
				"PumpSettings": state.PumpSettings,
				"Pump":         state.Pump,
				"Auth":         a,
			},
		}
		h.tsh.Broadcast(message.Target, message)
		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%s", name))
	}
}
