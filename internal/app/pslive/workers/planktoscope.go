// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func EstablishPlanktoscopeControllerConnections(
	ctx context.Context, is *instruments.Store, pco *planktoscope.Orchestrator,
) error {
	initialClients, err := is.GetControllersByProtocol(ctx, planktoscope.Protocol)
	if err != nil {
		return errors.Wrap(err, "couldn't determine which planktoscope controllers to connect to")
	}
	for _, client := range initialClients {
		if err := pco.Add(client.ID, client.URL); err != nil {
			return err
		}
	}

	return nil
}
