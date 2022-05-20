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

// Pump

const pumpPartial = "instruments/planktoscope/pump.partial.tmpl"

func replacePumpStream(
	id, controllerID int64, instrument instruments.Instrument, a auth.Auth, pc *planktoscope.Client,
) turbostreams.Message {
	state := pc.GetState()
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   fmt.Sprintf("/instruments/%d/controllers/%d/pump", id, controllerID),
		Template: pumpPartial,
		Data: map[string]interface{}{
			"Instrument":   instrument,
			"ControllerID": controllerID,
			"PumpSettings": state.PumpSettings,
			"Pump":         state.Pump,
			"Auth":         a,
		},
	}
}

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

func (h *Handlers) HandlePumpPub() turbostreams.HandlerFunc {
	t := pumpPartial
	h.r.MustHave(t)
	return func(c turbostreams.Context) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		controllerID, err := parseID(c.Param("controllerID"), "controller")
		if err != nil {
			return err
		}

		// Run queries
		instrument, err := h.is.GetInstrument(c.Context(), id)
		if err != nil {
			return err
		}
		pc, ok := h.pco.Get(id)
		if !ok {
			return errors.Errorf(
				"planktoscope client for controller %d on instrument %d not found", id, controllerID,
			)
		}

		// Publish on MQTT update
		for {
			ctx := c.Context()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-pc.PumpStateBroadcasted():
				if err := ctx.Err(); err != nil {
					// Context was also canceled and it should have priority
					return err
				}
				message := replacePumpStream(id, controllerID, instrument, auth.Auth{}, pc)
				c.Publish(message)
			}
		}
	}
}

func (h *Handlers) HandlePumpPost() auth.HTTPHandlerFunc {
	t := pumpPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		controllerID, err := parseID(c.Param("controllerID"), "controller")
		if err != nil {
			return err
		}

		// Run queries
		// FIXME: ensure that the controller belongs to the instrument and that the user is authorized!
		pc, ok := h.pco.Get(id)
		if !ok {
			return errors.Errorf(
				"planktoscope client for controller %d on instrument %d not found", id, controllerID,
			)
		}
		if err = handlePumpSettings(
			c.FormValue("pumping"), c.FormValue("direction"),
			c.FormValue("volume"), c.FormValue("flowrate"), pc,
		); err != nil {
			return err
		}

		// We rely on Turbo Streams over websockets, so we return an empty response here to avoid a race
		// condition of two Turbo Stream replace messages (where the one from this POST response could
		// be stale and overwrite a fresher message over websockets by arriving later).
		// FIXME: is there a cleaner way to avoid the race condition which would work even if the
		// WebSocket connection is misbehaving?
		if turbostreams.Accepted(c.Request().Header) {
			return h.r.TurboStream(c.Response())
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}
