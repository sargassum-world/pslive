package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentAutomationJobPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"automationJob",
		func(
			ctx context.Context, automationJobID instruments.AutomationJobID,
			enabled bool, params url.Values,
		) error {
			specificationType := params.Get("type")
			specification := params.Get("specification")
			if err := h.is.UpdateAutomationJob(ctx, instruments.AutomationJob{
				ID:            automationJobID,
				Enabled:       enabled,
				Type:          specificationType,
				Specification: specification,
			}); err != nil {
				return err
			}
			// Note: when we have other automation job types, we'll need to generalize this
			if !enabled {
				return h.ajo.Remove(ctx, automationJobID)
			}
			return h.ajo.Update(ctx, automationJobID, specificationType, specification)
		},
		func(ctx context.Context, automationJobID instruments.AutomationJobID) error {
			if err := h.is.DeleteAutomationJob(ctx, automationJobID); err != nil {
				return err
			}
			if err := h.ajo.Remove(ctx, automationJobID); err != nil {
				return err
			}
			return nil
		},
	)
}
