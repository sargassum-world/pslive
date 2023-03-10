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
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

// Pump

const pumpPartial = "instruments/planktoscope/pump.partial.tmpl"

func replacePumpStream(
	iid instruments.InstrumentID, cid instruments.ControllerID, a auth.Auth,
	pc *planktoscope.Client,
) turbostreams.Message {
	state := pc.GetState()
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   fmt.Sprintf("/instruments/%d/controllers/%d/pump", iid, cid),
		Template: pumpPartial,
		Data: map[string]interface{}{
			"InstrumentID": iid,
			"ControllerID": cid,
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
	return func(c *turbostreams.Context) error {
		// Parse params & run queries
		iid, cid, pc, err := getPlanktoscopeClientForPub(c, h.pco)
		if err != nil {
			return err
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
				// We insert an empty Auth object because the MSG handler will add the auth object for each
				// client
				message := replacePumpStream(iid, cid, auth.Auth{}, pc)
				c.Publish(message)
			}
		}
	}
}

type PlanktoscopePumpViewAuthz struct {
	Set bool
}

func getPlanktoscopePumpViewAuthz(
	ctx context.Context, iid instruments.InstrumentID, cid instruments.ControllerID,
	a auth.Auth, azc *auth.AuthzChecker,
) (authz PlanktoscopePumpViewAuthz, err error) {
	path := fmt.Sprintf("/instruments/%d/controllers/%d/pump", iid, cid)
	if authz.Set, err = azc.Allow(ctx, a, path, http.MethodPost, nil); err != nil {
		return PlanktoscopePumpViewAuthz{}, errors.Wrap(err, "couldn't check authz for setting pump")
	}
	return authz, nil
}

func (h *Handlers) ModifyPumpMsgData() handling.DataModifier {
	return func(
		ctx context.Context, a auth.Auth, data map[string]interface{},
	) (modifications map[string]interface{}, err error) {
		iid, cid, err := getIDsForModificationMiddleware(data)
		if err != nil {
			return nil, err
		}
		modifications = make(map[string]interface{})
		if modifications["Authorizations"], err = getPlanktoscopePumpViewAuthz(
			ctx, iid, cid, a, h.azc,
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't check authz for pump of controller %d of instrument %d", cid, iid,
			)
		}
		return modifications, nil
	}
}

func (h *Handlers) HandlePumpPost() auth.HTTPHandlerFunc {
	t := pumpPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		cid, err := parseID[instruments.ControllerID](c.Param("controllerID"), "controller")
		if err != nil {
			return err
		}

		// Run queries
		pc, ok := h.pco.Get(planktoscope.ClientID(cid))
		if !ok {
			return errors.Errorf(
				"planktoscope client for controller %d on instrument %d not found for pump post", cid, iid,
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
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}

// Camera

const cameraPartial = "instruments/planktoscope/camera.partial.tmpl"

func replaceCameraStream(
	iid instruments.InstrumentID, cid instruments.ControllerID, a auth.Auth,
	pc *planktoscope.Client,
) turbostreams.Message {
	state := pc.GetState()
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   fmt.Sprintf("/instruments/%d/controllers/%d/camera", iid, cid),
		Template: cameraPartial,
		Data: map[string]interface{}{
			"InstrumentID":   iid,
			"ControllerID":   cid,
			"CameraSettings": state.CameraSettings,
			"Auth":           a,
		},
	}
}

func handleCameraSettings(
	isoRaw, shutterSpeedRaw,
	autoWhiteBalanceRaw, whiteBalanceRedGainRaw, whiteBalanceBlueGainRaw string,
	pc *planktoscope.Client,
) (err error) {
	var token mqtt.Token
	// TODO: use echo's request binding functionality instead of strconv.ParseFloat
	// TODO: perform input validation and handle invalid inputs
	const uintBase = 10
	const uintWidth = 64
	iso, err := strconv.ParseUint(isoRaw, uintBase, uintWidth)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse iso"))
	}
	shutterSpeed, err := strconv.ParseUint(shutterSpeedRaw, uintBase, uintWidth)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "couldn't parse shutter speed"))
	}

	const floatWidth = 64
	autoWhiteBalance := strings.ToLower(autoWhiteBalanceRaw) == "true"
	whiteBalanceRedGain, err := strconv.ParseFloat(whiteBalanceRedGainRaw, floatWidth)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(
			err, "couldn't parse white balance red gain",
		))
	}
	whiteBalanceBlueGain, err := strconv.ParseFloat(whiteBalanceBlueGainRaw, floatWidth)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(
			err, "couldn't parse white balance blue gain",
		))
	}

	if token, err = pc.SetCamera(
		iso, shutterSpeed, autoWhiteBalance, whiteBalanceRedGain, whiteBalanceBlueGain,
	); err != nil {
		return err
	}

	stateUpdated := pc.CameraStateBroadcasted()
	// TODO: instead of waiting forever, have a timeout before redirecting and displaying a
	// warning message that we haven't heard any camera settings updates from the planktoscope.
	if token.Wait(); token.Error() != nil {
		return token.Error()
	}
	<-stateUpdated
	return nil
}

func (h *Handlers) HandleCameraPub() turbostreams.HandlerFunc {
	t := cameraPartial
	h.r.MustHave(t)
	return func(c *turbostreams.Context) error {
		// Parse params & run queries
		iid, cid, pc, err := getPlanktoscopeClientForPub(c, h.pco)
		if err != nil {
			return err
		}

		// Publish on MQTT update
		for {
			ctx := c.Context()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-pc.CameraStateBroadcasted():
				if err := ctx.Err(); err != nil {
					// Context was also canceled and it should have priority
					return err
				}
				// We insert an empty Auth object because the MSG handler will add the auth object for each
				// client
				message := replaceCameraStream(iid, cid, auth.Auth{}, pc)
				c.Publish(message)
			}
		}
	}
}

type PlanktoscopeCameraViewAuthz struct {
	Set bool
}

func getPlanktoscopeCameraViewAuthz(
	ctx context.Context, iid instruments.InstrumentID, cid instruments.ControllerID,
	a auth.Auth, azc *auth.AuthzChecker,
) (authz PlanktoscopeCameraViewAuthz, err error) {
	path := fmt.Sprintf("/instruments/%d/controllers/%d/camera", iid, cid)
	if authz.Set, err = azc.Allow(ctx, a, path, http.MethodPost, nil); err != nil {
		return PlanktoscopeCameraViewAuthz{}, errors.Wrap(
			err, "couldn't check authz for setting camera",
		)
	}
	return authz, nil
}

func (h *Handlers) ModifyCameraMsgData() handling.DataModifier {
	return func(
		ctx context.Context, a auth.Auth, data map[string]interface{},
	) (modifications map[string]interface{}, err error) {
		iid, cid, err := getIDsForModificationMiddleware(data)
		if err != nil {
			return nil, err
		}
		modifications = make(map[string]interface{})
		if modifications["Authorizations"], err = getPlanktoscopeCameraViewAuthz(
			ctx, iid, cid, a, h.azc,
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't check authz for camera of controller %d of instrument %d", cid, iid,
			)
		}
		return modifications, nil
	}
}

func (h *Handlers) HandleCameraPost() auth.HTTPHandlerFunc {
	t := cameraPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		cid, err := parseID[instruments.ControllerID](c.Param("controllerID"), "controller")
		if err != nil {
			return err
		}

		// Run queries
		pc, ok := h.pco.Get(planktoscope.ClientID(cid))
		if !ok {
			return errors.Errorf(
				"planktoscope client for controller %d on instrument %d not found for camera post",
				cid, iid,
			)
		}
		if err = handleCameraSettings(
			c.FormValue("iso"), c.FormValue("shutter-speed"),
			c.FormValue("awb"), c.FormValue("wb-red"), c.FormValue("wb-blue"), pc,
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
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}

// Controller

type PlanktoscopeControllerViewAuthz struct {
	Pump   PlanktoscopePumpViewAuthz
	Camera PlanktoscopeCameraViewAuthz
}

func getPlanktoscopeControllerViewAuthz(
	ctx context.Context, iid instruments.InstrumentID, cid instruments.ControllerID,
	a auth.Auth, azc *auth.AuthzChecker,
) (authz PlanktoscopeControllerViewAuthz, err error) {
	if authz.Pump, err = getPlanktoscopePumpViewAuthz(
		ctx, iid, cid, a, azc,
	); err != nil {
		return PlanktoscopeControllerViewAuthz{}, errors.Wrap(err, "couldn't check authz for pump")
	}
	if authz.Camera, err = getPlanktoscopeCameraViewAuthz(
		ctx, iid, cid, a, azc,
	); err != nil {
		return PlanktoscopeControllerViewAuthz{}, errors.Wrap(err, "couldn't check authz for camera")
	}
	return authz, nil
}

func getIDsForModificationMiddleware(
	data map[string]interface{},
) (iid instruments.InstrumentID, cid instruments.ControllerID, err error) {
	rawIID, ok := data["InstrumentID"]
	if !ok {
		return 0, 0, errors.New(
			"couldn't find instrument id from turbostreams message data to check authorizations",
		)
	}
	iid, ok = rawIID.(instruments.InstrumentID)
	if !ok {
		return 0, 0, errors.Errorf(
			"instrument id has unexpected type %T in turbostreams message data for checking authorization",
			rawIID,
		)
	}
	rawCID, ok := data["ControllerID"]
	if !ok {
		return 0, 0, errors.Errorf(
			"couldn't find controller id for instrument %d from turbostreams message data to check authorizations",
			iid,
		)
	}
	cid, ok = rawCID.(instruments.ControllerID)
	if !ok {
		return 0, 0, errors.Errorf(
			"controller id has unexpected type %T in turbostreams message data for checking authorization",
			rawCID,
		)
	}
	return iid, cid, nil
}

func getPlanktoscopeClientForPub(
	c *turbostreams.Context, pco *planktoscope.Orchestrator,
) (
	iid instruments.InstrumentID, cid instruments.ControllerID, client *planktoscope.Client,
	err error,
) {
	// Parse params
	iid, err = parseID[instruments.InstrumentID](c.Param("id"), "instrument")
	if err != nil {
		return 0, 0, nil, err
	}
	cid, err = parseID[instruments.ControllerID](c.Param("controllerID"), "controller")
	if err != nil {
		return 0, 0, nil, err
	}

	// Run queries
	pc, ok := pco.Get(planktoscope.ClientID(cid))
	if !ok {
		return 0, 0, nil, errors.Errorf(
			"planktoscope client for controller %d on instrument %d not found for pub",
			cid, iid,
		)
	}
	return iid, cid, pc, nil
}
