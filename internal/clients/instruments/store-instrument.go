package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

//go:embed queries/insert-instrument.sql
var rawInsertInstrumentQuery string
var insertInstrumentQuery string = strings.TrimSpace(rawInsertInstrumentQuery)

func (s *Store) AddInstrument(ctx context.Context, i Instrument) (instrumentID int64, err error) {
	rowID, err := s.db.ExecuteInsertionForID(ctx, insertInstrumentQuery, i.newInsertion())
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't add instrument with admin %s", i.AdminID)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation
	return rowID, err
}

//go:embed queries/update-instrument-name.sql
var rawUpdateInstrumentNameQuery string
var updateInstrumentNameQuery string = strings.TrimSpace(rawUpdateInstrumentNameQuery)

func (s *Store) UpdateInstrumentName(
	ctx context.Context, id InstrumentID, name string,
) (err error) {
	return errors.Wrapf(
		s.db.ExecuteUpdate(ctx, updateInstrumentNameQuery, Instrument{
			ID:   id,
			Name: name,
		}.newNameUpdate()),
		"couldn't update name of instrument %d", id,
	)
}

//go:embed queries/update-instrument-description.sql
var rawUpdateInstrumentDescriptionQuery string
var updateInstrumentDescriptionQuery string = strings.TrimSpace(rawUpdateInstrumentDescriptionQuery)

func (s *Store) UpdateInstrumentDescription(
	ctx context.Context, id InstrumentID, description string,
) (err error) {
	return errors.Wrapf(
		s.db.ExecuteUpdate(ctx, updateInstrumentDescriptionQuery, Instrument{
			ID:          id,
			Description: description,
		}.newDescriptionUpdate()),
		"couldn't update description of instrument %d", id,
	)
}

//go:embed queries/delete-instrument.sql
var rawDeleteInstrumentQuery string
var deleteInstrumentQuery string = strings.TrimSpace(rawDeleteInstrumentQuery)

func (s *Store) DeleteInstrument(ctx context.Context, id InstrumentID) (err error) {
	return errors.Wrapf(
		s.db.ExecuteDelete(ctx, deleteInstrumentQuery, Instrument{ID: id}.newDelete()),
		"couldn't delete instrument %d", id,
	)
}

//go:embed queries/select-instrument.sql
var rawSelectInstrumentQuery string
var selectInstrumentQuery string = strings.TrimSpace(rawSelectInstrumentQuery)

func (s *Store) GetInstrument(ctx context.Context, id InstrumentID) (i Instrument, err error) {
	sel := newInstrumentsSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectInstrumentQuery, newInstrumentSelection(id), sel.Step,
	); err != nil {
		return Instrument{}, errors.Wrapf(err, "couldn't get instrument with id %d", id)
	}
	instruments := sel.Instruments()
	if len(instruments) == 0 {
		return Instrument{}, errors.Errorf("couldn't get non-existent instrument with id %d", id)
	}
	return instruments[0], nil
}
