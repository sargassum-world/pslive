package planktoscope

import (
	"encoding/json"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

func (c *Client) ImagerStateBroadcasted() <-chan struct{} {
	return c.imagerB.Broadcasted()
}

// Receive Updates

func (c *Client) updateImagerState(newState Imager) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.imager = newState
	c.imagerB.BroadcastNext()
}

func (c *Client) handleImagerStatusUpdate(_ string, rawPayload []byte) error {
	type ImagerStatus struct {
		Status   string  `json:"status"`
		Duration float64 `json:"duration"`
	}
	var payload ImagerStatus
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newState := Imager{
		StateKnown: true,
	}
	switch status := payload.Status; status {
	default:
		// TODO: write the status to the imager state for display in the GUI
		c.Logger.Infof("unknown status %s", status)
		return nil
	case "Camera settings updated":
		return nil
	case "Started":
		newState.Imaging = true
		newState.Start = time.Now()
	case "Interrupted":
		newState.Imaging = false
	case "Done":
		newState.Imaging = false
	}

	// Commit changes
	c.updateImagerState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newState)
	return nil
}

func (c *Client) updateImagerSettings(newSettings ImagerSettings) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.imagerSettings = newSettings
	c.imagerB.BroadcastNext()
}

const (
	imageCommand = "image"
	stopCommand  = "stop"
)

func (c *Client) handleImagerImagingUpdate(_ string, rawPayload []byte) error {
	type ImageCommand struct {
		Action     string  `json:"action"`
		Direction  string  `json:"pump_direction,omitempty"`
		StepVolume float64 `json:"volume,omitempty"`
		StepDelay  float64 `json:"sleep,omitempty"`
		Steps      uint64  `json:"nb_frame,omitempty"`
	}
	var payload ImageCommand
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newSettings := ImagerSettings{}
	switch action := payload.Action; action {
	default:
		return errors.Errorf("unknown action %s", action)
	case stopCommand:
		// No settings to update
		break
	case imageCommand:
		// Parse direction
		switch direction := payload.Direction; direction {
		default:
			return errors.Errorf("unknown direction %s", direction)
		case forwardDirection:
			newSettings.Forward = true
		case backwardDirection:
			newSettings.Forward = false
		}

		newSettings.StepVolume = payload.StepVolume
		newSettings.StepDelay = payload.StepDelay
		newSettings.Steps = payload.Steps

		// Commit changes
		c.updateImagerSettings(newSettings)
		c.Logger.Debugf("%s: %+v", c.Config.URL, newSettings)
	}
	return nil
}

func (c *Client) handleImagerUpdate(topic string, rawPayload []byte) error {
	type ImagerBaseCommand struct {
		Action string `json:"action"`
	}
	var basePayload ImagerBaseCommand
	if err := json.Unmarshal(rawPayload, &basePayload); err != nil {
		return errors.Wrapf(err, "unparseable base payload")
	}
	broker := c.Config.URL
	switch action := basePayload.Action; action {
	default:
		var payload interface{}
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			c.Logger.Errorf("%s/%s: unknown payload %s", broker, topic, rawPayload)
			return nil
		}
		c.Logger.Infof("%s/%s: %v", broker, topic, payload)
	case stopCommand:
		// No settings to update
		break
	case "settings":
		if err := c.handleCameraSettingsUpdate(topic, rawPayload); err != nil {
			return errors.Wrap(err, "invalid camera settings command")
		}
	case imageCommand:
		if err := c.handleImagerImagingUpdate(topic, rawPayload); err != nil {
			return errors.Wrap(err, "invalid imager config update command")
		}
	}
	return nil
}

// Send Commands

func (c *Client) StopImaging() (mqtt.Token, error) {
	command := struct {
		Action string `json:"action"`
	}{
		Action: stopCommand,
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	token := c.MQTT.Publish("imager/image", mqttAtLeastOnce, false, marshaled)
	return token, nil
}

func (c *Client) StartImaging(
	forward bool, stepVolume, stepDelay float64, steps uint64,
) (mqtt.Token, error) {
	command := struct {
		Action     string  `json:"action"`
		Direction  string  `json:"pump_direction"`
		StepVolume float64 `json:"volume"`
		StepDelay  float64 `json:"sleep"`
		Steps      uint64  `json:"nb_frame"`
	}{
		Action:     imageCommand,
		StepVolume: stepVolume,
		StepDelay:  stepDelay,
		Steps:      steps,
	}
	if forward {
		command.Direction = forwardDirection
	} else {
		command.Direction = backwardDirection
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.imagerSettings.Forward = forward
	c.imagerSettings.StepVolume = stepVolume
	c.imagerSettings.StepDelay = stepDelay
	c.imagerSettings.Steps = steps

	token := c.MQTT.Publish("imager/image", mqttExactlyOnce, false, marshaled)
	return token, nil
}
