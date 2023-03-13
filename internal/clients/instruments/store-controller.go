package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/insert-controller.sql
var rawInsertControllerQuery string
var insertControllerQuery string = strings.TrimSpace(rawInsertControllerQuery)

func (s *Store) AddController(ctx context.Context, c Controller) (controllerID int64, err error) {
	rowID, err := s.db.ExecuteInsertionForID(ctx, insertControllerQuery, c.newInsertion())
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't add controller for instrument %d", c.InstrumentID)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation
	return rowID, err
}

//go:embed queries/update-controller.sql
var rawUpdateControllerQuery string
var updateControllerQuery string = strings.TrimSpace(rawUpdateControllerQuery)

func (s *Store) UpdateController(ctx context.Context, c Controller) (err error) {
	return errors.Wrapf(
		s.db.ExecuteUpdate(ctx, updateControllerQuery, c.newUpdate()),
		"couldn't update controller %d", c.ID,
	)
}

//go:embed queries/delete-controller.sql
var rawDeleteControllerQuery string
var deleteControllerQuery string = strings.TrimSpace(rawDeleteControllerQuery)

func (s *Store) DeleteController(ctx context.Context, id ControllerID) (err error) {
	return errors.Wrapf(
		s.db.ExecuteDelete(ctx, deleteControllerQuery, Controller{ID: id}.newDelete()),
		"couldn't delete controller %d", id,
	)
}
