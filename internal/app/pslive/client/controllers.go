package client

import (
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func NewPlanktoScopeControllerActionRunnerGetter(
	o *planktoscope.Orchestrator,
) instruments.ControllerActionRunnerGetter {
	return func(id instruments.ControllerID) (a instruments.ControllerActionRunner, ok bool) {
		return o.Get(planktoscope.ClientID(id))
	}
}
