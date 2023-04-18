package planktoscope

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

func (c *Client) PumpStateBroadcasted() <-chan struct{} {
	return c.pumpB.Broadcasted()
}

// Receive Updates

func (c *Client) updatePumpState(newState Pump) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.pump = newState
	c.pumpB.BroadcastNext()
}

func (c *Client) handlePumpStatusUpdate(_ string, rawPayload []byte) error {
	type PumpStatus struct {
		Status   string  `json:"status"`
		Duration float64 `json:"duration"`
	}
	var payload PumpStatus
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newState := Pump{
		StateKnown: true,
		Start:      time.Now(),
	}
	switch status := payload.Status; status {
	default:
		// TODO: write the status to the imager state for display in the GUI
		return errors.Errorf("unknown status %s", status)
	case "Started":
		newState.Pumping = true
		newState.Duration = time.Duration(payload.Duration) * time.Second
	case "Interrupted":
		newState.Pumping = false
		newState.Duration = 0
	case "Done":
		newState.Pumping = false
		newState.Duration = 0
	}
	newState.Deadline = newState.Start.Add(newState.Duration)

	// Commit changes
	c.updatePumpState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newState)
	return nil
}

func (c *Client) updatePumpSettings(newSettings PumpSettings) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.pumpSettings = newSettings
	c.pumpB.BroadcastNext()
}

func parseFloat(n interface{}) (float64, error) {
	switch number := n.(type) {
	default:
		return 0, errors.Errorf("unknown float type %T", number)
	case float64:
		return number, nil
	case string:
		const floatWidth = 64
		parsed, err := strconv.ParseFloat(number, floatWidth)
		return parsed, errors.Wrapf(err, "couldn't parse number %s", number)
	}
}

const (
	forwardDirection  = "FORWARD"
	backwardDirection = "BACKWARD"
)

func (c *Client) handlePumpActuatorUpdate(_ string, rawPayload []byte) error {
	type PumpCommand struct {
		Action    string `json:"action"`
		Direction string `json:"direction,omitempty"`
		// The Node-Red dashboard may send volume and flowrate as either string or number
		Volume   interface{} `json:"volume,omitempty"`
		Flowrate interface{} `json:"flowrate,omitempty"`
	}
	var payload PumpCommand
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newSettings := PumpSettings{}
	switch action := payload.Action; action {
	default:
		return errors.Errorf("unknown action %s", action)
	case "stop":
		// No settings to update
		break
	case "move":
		// Parse direction
		switch direction := payload.Direction; direction {
		default:
			return errors.Errorf("unknown direction %s", direction)
		case forwardDirection:
			newSettings.Forward = true
		case backwardDirection:
			newSettings.Forward = false
		}

		// Parse volume
		volume, err := parseFloat(payload.Volume)
		if err != nil {
			return errors.Wrap(err, "couldn't parse new pump volume setting")
		}
		newSettings.Volume = volume

		// Parse flowrate
		flowrate, err := parseFloat(payload.Flowrate)
		if err != nil {
			return errors.Wrap(err, "couldn't parse new pump flowrate setting")
		}
		newSettings.Flowrate = flowrate

		// Commit changes
		c.updatePumpSettings(newSettings)
		c.Logger.Debugf("%s: %+v", c.Config.URL, newSettings)
	}
	return nil
}

// Send Commands

func (c *Client) StopPump() (mqtt.Token, error) {
	command := struct {
		Action string `json:"action"`
	}{
		Action: "stop",
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	token := c.MQTT.Publish("actuator/pump", mqttAtLeastOnce, false, marshaled)
	return token, nil
}

func (c *Client) StartPump(forward bool, volume, flowrate float64) (mqtt.Token, error) {
	command := struct {
		Action    string  `json:"action"`
		Direction string  `json:"direction"`
		Volume    float64 `json:"volume"`
		Flowrate  float64 `json:"flowrate"`
	}{
		Action:   "move",
		Volume:   volume,
		Flowrate: flowrate,
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

	c.pumpSettings.Forward = forward
	c.pumpSettings.Volume = volume
	c.pumpSettings.Flowrate = flowrate

	token := c.MQTT.Publish("actuator/pump", mqttExactlyOnce, false, marshaled)
	return token, nil
}
