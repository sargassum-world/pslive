package planktoscope

import (
	"encoding/json"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

type Pump struct {
	StateKnown bool
	Pumping    bool
	Start      time.Time
	Duration   time.Duration
	Deadline   time.Time
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
	c.updatePumpState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.Broker().String(), newState)
	return nil
}

func (c *Client) updatePumpState(newState Pump) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.pump = newState
	c.pumpB.BroadcastNext()
	// TODO: push updated state to clients
}

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

type PumpSettings struct {
	Forward  bool
	Volume   float64
	Flowrate float64
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
		command.Direction = "FORWARD"
	} else {
		command.Direction = "BACKWARD"
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
	// TODO: push updated settings to clients

	token := c.MQTT.Publish("actuator/pump", mqttExactlyOnce, false, marshaled)
	return token, nil
}

func (c *Client) PumpStateBroadcasted() <-chan struct{} {
	return c.pumpB.Broadcasted()
}
