// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func EstablishPlanktoscopeControllerConnection(c *planktoscope.Client) error {
	if err := c.EstablishConnection(); err != nil {
		return errors.Wrap(err, "couldn't establish connection to the Planktoscope controller")
	}
	return nil
}

func EstablishPlanktoscopeControllerConnections(
	ctx context.Context, pcs map[string]*planktoscope.Client,
) error {
	// TODO: persistently retry failed connections
	eg, _ := errgroup.WithContext(ctx)
	for _, pc := range pcs {
		eg.Go(func(c *planktoscope.Client) func() error {
			return func() error {
				return EstablishPlanktoscopeControllerConnection(c)
			}
		}(pc))
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
