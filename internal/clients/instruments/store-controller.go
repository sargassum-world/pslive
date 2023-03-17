package instruments

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed queries/insert-controller.sql
var rawInsertControllerQuery string
var insertControllerQuery string = strings.TrimSpace(rawInsertControllerQuery)

func (s *Store) AddController(ctx context.Context, c Controller) (cid ControllerID, err error) {
	return executeComponentInsert[ControllerID](ctx, insertControllerQuery, c, s.db)
}

//go:embed queries/update-controller.sql
var rawUpdateControllerQuery string
var updateControllerQuery string = strings.TrimSpace(rawUpdateControllerQuery)

func (s *Store) UpdateController(ctx context.Context, c Controller) (err error) {
	return executeUpdate[ControllerID](ctx, updateControllerQuery, c, s.db)
}

//go:embed queries/delete-controller.sql
var rawDeleteControllerQuery string
var deleteControllerQuery string = strings.TrimSpace(rawDeleteControllerQuery)

func (s *Store) DeleteController(ctx context.Context, id ControllerID) (err error) {
	return executeDelete[ControllerID](ctx, deleteControllerQuery, Controller{ID: id}, s.db)
}
