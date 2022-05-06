package instruments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type InstrumentViewData struct {
	Instrument       instruments.Instrument
	Controller       planktoscope.Planktoscope
	KnownViewers     []presence.User
	AnonymousViewers []string
	AdminIdentifier  string
	ChatMessages     []chat.Message
}

func getInstrumentViewData(
	ctx context.Context, name string,
	oc *ory.Client, ic *instruments.Client, pcs map[string]*planktoscope.Client,
	ps *presence.Store, cs *chat.Store,
) (*InstrumentViewData, error) {
	instrument, err := ic.FindInstrument(name)
	if err != nil {
		return nil, err
	}
	if instrument == nil {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %s not found", name),
		)
	}

	adminIdentifier, err := oc.GetIdentifier(ctx, instrument.Administrator)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't look up admin identifier for instrument %s", name)
	}

	pc, ok := pcs[instrument.Controller]
	if !ok {
		return nil, errors.Errorf("planktoscope client for instrument %s not found", name)
	}
	known, anonymous := ps.List("/instruments/" + name + "/users")
	messages, err := cs.GetMessagesByTopic(
		ctx, "/instruments/" + name + "/chat/messages", chat.DefaultMessagesLimit,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get chat messages for instrument %s", name)
	}
	return &InstrumentViewData{
		Instrument:       *instrument,
		Controller:       pc.GetState(),
		AdminIdentifier:  adminIdentifier,
		KnownViewers:     known,
		AnonymousViewers: anonymous,
		ChatMessages:     messages,
	}, nil
}

func (h *Handlers) HandleInstrumentGet() auth.HTTPHandlerFunc {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		instrumentViewData, err := getInstrumentViewData(
			c.Request().Context(), name, h.oc, h.ic, h.pcs, h.ps, h.cs,
		)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *instrumentViewData, a)
	}
}

// Pumping

const pumpPartial = "instruments/planktoscope/pump.partial.tmpl"

func replacePumpStream(
	name string, instrument *instruments.Instrument, a auth.Auth, pc *planktoscope.Client,
) turbostreams.Message {
	state := pc.GetState()
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   "/instruments/" + name + "/controller/pump",
		Template: pumpPartial,
		Data: map[string]interface{}{
			"Instrument":   instrument,
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
				message := replacePumpStream(name, instrument, auth.Auth{}, pc)
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

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			message := replacePumpStream(name, instrument, a, pc)
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/instruments/"+name)
	}
}
