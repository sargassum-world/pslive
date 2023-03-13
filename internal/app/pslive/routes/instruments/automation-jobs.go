package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentAutomationJobsPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, iid instruments.InstrumentID,
			enabled bool, name, description string, params url.Values,
		) error {
			specificationType := params.Get("type")
			specification := params.Get("specification")
			automationJobID, err := h.is.AddAutomationJob(ctx, instruments.AutomationJob{
				InstrumentID:  iid,
				Enabled:       enabled,
				Name:          name,
				Description:   description,
				Type:          specificationType,
				Specification: specification,
			})
			if err != nil {
				return err
			}
			if !enabled {
				return nil
			}
			return h.ajo.Add(automationJobID, specificationType, specification)
		},
	)
}
