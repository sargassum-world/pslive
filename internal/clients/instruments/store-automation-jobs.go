package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/select-enabled-automation-jobs.sql
var rawSelectEnabledAutomationJobsQuery string

var selectEnabledAutomationJobsQuery string = strings.TrimSpace(
	rawSelectEnabledAutomationJobsQuery,
)

func (s *Store) GetEnabledAutomationJobs(ctx context.Context) (jobs []AutomationJob, err error) {
	sel := newAutomationJobsSelector()
	if err = s.db.ExecuteSelection(ctx, selectEnabledAutomationJobsQuery, nil, sel.Step); err != nil {
		return nil, errors.Wrap(err, "couldn't get enabled automation jobs")
	}
	return sel.AutomationJobs(), nil
}
