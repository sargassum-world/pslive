// Package workers provides functionality which runs independently of request servicing.
package workers

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func EstablishPlanktoscopeControllerConnection(c *planktoscope.Client) error {
	if err := c.EstablishConnection(); err != nil {
		return errors.Wrap(err, "couldn't establish connection to the Planktoscope controller")
	}
	return nil
}
