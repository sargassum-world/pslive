package instruments

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed queries/insert-automation-job.sql
var rawInsertAutomationJobQuery string
var insertAutomationJobQuery string = strings.TrimSpace(rawInsertAutomationJobQuery)

func (s *Store) AddAutomationJob(
	ctx context.Context, c AutomationJob,
) (automationJobID AutomationJobID, err error) {
	return executeComponentInsert[AutomationJobID](ctx, insertAutomationJobQuery, c, s.db)
}

//go:embed queries/update-automation-job.sql
var rawUpdateAutomationJobQuery string
var updateAutomationJobQuery string = strings.TrimSpace(rawUpdateAutomationJobQuery)

func (s *Store) UpdateAutomationJob(ctx context.Context, c AutomationJob) (err error) {
	return executeUpdate[AutomationJobID](ctx, updateAutomationJobQuery, c, s.db)
}

//go:embed queries/delete-automation-job.sql
var rawDeleteAutomationJobQuery string
var deleteAutomationJobQuery string = strings.TrimSpace(rawDeleteAutomationJobQuery)

func (s *Store) DeleteAutomationJob(ctx context.Context, id AutomationJobID) (err error) {
	return executeDelete[AutomationJobID](ctx, deleteAutomationJobQuery, AutomationJob{ID: id}, s.db)
}
