package planktoscope

import (
	"encoding/json"
	"time"

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
		Duration float32 `json:"duration"`
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
	}
	newState.Deadline = newState.Start.Add(newState.Duration)
	c.updatePumpState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.Broker().String(), newState)
	return nil
}

func (c *Client) updatePumpState(newState Pump) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	c.pump = newState
	// TODO: notify clients that the state has updated
}
