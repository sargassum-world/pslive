package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/insert-automation-job.sql
var rawInsertAutomationJobQuery string
var insertAutomationJobQuery string = strings.TrimSpace(rawInsertAutomationJobQuery)

func (s *Store) AddAutomationJob(
	ctx context.Context, c AutomationJob,
) (automationJobID AutomationJobID, err error) {
	rowID, err := s.db.ExecuteInsertionForID(ctx, insertAutomationJobQuery, c.newInsertion())
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't add automation job for instrument %d", c.InstrumentID)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation
	return AutomationJobID(rowID), err
}

//go:embed queries/update-automation-job.sql
var rawUpdateAutomationJobQuery string
var updateAutomationJobQuery string = strings.TrimSpace(rawUpdateAutomationJobQuery)

func (s *Store) UpdateAutomationJob(ctx context.Context, c AutomationJob) (err error) {
	return errors.Wrapf(
		s.db.ExecuteUpdate(ctx, updateAutomationJobQuery, c.newUpdate()),
		"couldn't update automation job %d", c.ID,
	)
}

//go:embed queries/delete-automation-job.sql
var rawDeleteAutomationJobQuery string
var deleteAutomationJobQuery string = strings.TrimSpace(rawDeleteAutomationJobQuery)

func (s *Store) DeleteAutomationJob(ctx context.Context, id AutomationJobID) (err error) {
	return errors.Wrapf(
		s.db.ExecuteDelete(ctx, deleteAutomationJobQuery, AutomationJob{ID: id}.newDelete()),
		"couldn't delete automation job %d", id,
	)
}
