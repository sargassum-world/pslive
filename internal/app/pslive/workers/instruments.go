package workers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func StartInstrumentJobs(
	ctx context.Context, is *instruments.Store, ajo *instruments.JobOrchestrator,
) error {
	initialJobs, err := is.GetEnabledAutomationJobs(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't determine which automation jobs to start")
	}
	for _, job := range initialJobs {
		if err := ajo.Add(job.ID, job.InstrumentID, job.Name, job.Type, job.Specification); err != nil {
			return err
		}
	}

	return nil
}
