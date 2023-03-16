package planktoscope

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"
)

// Pump Actions

type PlanktoscopePumpParams struct {
	Forward  bool    `hcl:"forward"`
	Volume   float64 `hcl:"volume"`
	Flowrate float64 `hcl:"flowrate"`
}

func (c *Client) RunPumpAction(ctx context.Context, p PlanktoscopePumpParams) error {
	token, err := c.StartPump(p.Forward, p.Volume, p.Flowrate)
	if err != nil {
		return errors.Wrap(err, "couldn't send command to start the pump")
	}
	stateUpdated := c.PumpStateBroadcasted()
	// TODO: instead of always waiting forever, have an action-configured optional timeout before
	// returning an error that we haven't heard any pump updates from the planktoscope.
	if token.Wait(); token.Error() != nil {
		return token.Error()
	}
	<-stateUpdated
	return nil
}

func (c *Client) RunStopPumpAction(ctx context.Context) error {
	token, err := c.StopPump()
	if err != nil {
		return errors.Wrap(err, "couldn't send command to stop the pump")
	}
	stateUpdated := c.PumpStateBroadcasted()
	// TODO: instead of always waiting forever, have an action-configured optional timeout before
	// returning an error that we haven't heard any pump updates from the planktoscope.
	if token.Wait(); token.Error() != nil {
		return token.Error()
	}
	<-stateUpdated
	return nil
}

// Controller Action

func (c *Client) RunControllerAction(ctx context.Context, command string, params hcl.Body) error {
	switch command {
	default:
		return errors.Errorf("unrecognized planktoscope controller command %s", command)
	case "pump":
		var p PlanktoscopePumpParams
		if err := gohcl.DecodeBody(params, nil, &p); err != nil {
			return errors.Wrapf(
				err, "couldn't decode params of planktoscope controller command %s", command,
			)
		}
		return c.RunPumpAction(ctx, p)
	case "stop-pump":
		return c.RunStopPumpAction(ctx)
	}
}
