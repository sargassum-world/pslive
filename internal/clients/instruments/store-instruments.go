package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/select-instruments.sql
var rawSelectInstrumentsQuery string
var selectInstrumentsQuery string = strings.TrimSpace(rawSelectInstrumentsQuery)

func (s *Store) GetInstruments(ctx context.Context) (instruments []Instrument, err error) {
	sel := newInstrumentsSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectInstrumentsQuery, newInstrumentsSelection(), sel.Step,
	); err != nil {
		return nil, errors.Wrapf(err, "couldn't get instruments")
	}
	return sel.Instruments(), nil
}

//go:embed queries/select-instruments-by-admin-id.sql
var rawSelectInstrumentsByAdminIDQuery string
var selectInstrumentsByAdminIDQuery string = strings.TrimSpace(rawSelectInstrumentsByAdminIDQuery)

func (s *Store) GetInstrumentsByAdminID(
	ctx context.Context, adminID AdminID,
) (instruments []Instrument, err error) {
	sel := newInstrumentsSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectInstrumentsByAdminIDQuery, Instrument{AdminID: adminID}.newAdminIDSelection(),
		sel.Step,
	); err != nil {
		return nil, errors.Wrapf(err, "couldn't get instruments with admin id %s", adminID)
	}
	return sel.Instruments(), nil
}
