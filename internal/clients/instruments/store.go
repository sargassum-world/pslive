// Package instruments provides a high-level client for management of imaging instruments
package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/database"
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{
		db: db,
	}
}

// Camera

//go:embed queries/insert-camera.sql
var rawInsertCameraQuery string
var insertCameraQuery string = strings.TrimSpace(rawInsertCameraQuery)

func (s *Store) AddCamera(ctx context.Context, c Camera) (cameraID int64, err error) {
	rowID, err := s.db.ExecuteInsertionForID(ctx, insertCameraQuery, c.newInsertion())
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't add camera for instrument %d", c.InstrumentID)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation
	return rowID, err
}

//go:embed queries/update-camera.sql
var rawUpdateCameraQuery string
var updateCameraQuery string = strings.TrimSpace(rawUpdateCameraQuery)

func (s *Store) UpdateCamera(ctx context.Context, c Camera) (err error) {
	return errors.Wrapf(
		s.db.ExecuteUpdate(ctx, updateCameraQuery, c.newUpdate()),
		"couldn't update camera %d", c.ID,
	)
}

//go:embed queries/delete-camera.sql
var rawDeleteCameraQuery string
var deleteCameraQuery string = strings.TrimSpace(rawDeleteCameraQuery)

func (s *Store) DeleteCamera(ctx context.Context, id CameraID) (err error) {
	return errors.Wrapf(
		s.db.ExecuteDelete(ctx, deleteCameraQuery, Camera{ID: id}.newDelete()),
		"couldn't delete camera %d", id,
	)
}

//go:embed queries/select-camera.sql
var rawSelectCameraQuery string
var selectCameraQuery string = strings.TrimSpace(rawSelectCameraQuery)

func (s *Store) GetCamera(ctx context.Context, id CameraID) (i Camera, err error) {
	sel := newCamerasSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectCameraQuery, newCameraSelection(id), sel.Step,
	); err != nil {
		return Camera{}, errors.Wrapf(err, "couldn't get camera with id %d", id)
	}
	cameras := sel.Cameras()
	if len(cameras) == 0 {
		return Camera{}, errors.Errorf("couldn't get non-existent camera with id %d", id)
	}
	return cameras[0], nil
}

// Controller

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

// Controllers

//go:embed queries/select-controllers-by-protocol.sql
var rawSelectControllersByProtocolQuery string
var selectControllersByProtocolQuery string = strings.TrimSpace(rawSelectControllersByProtocolQuery)

func (s *Store) GetControllersByProtocol(
	ctx context.Context, protocol string,
) (controllers []Controller, err error) {
	sel := newControllersSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectControllersByProtocolQuery, Controller{Protocol: protocol}.newProtocolSelection(),
		sel.Step,
	); err != nil {
		return nil, errors.Wrapf(err, "couldn't get controllers with protocol %s", protocol)
	}
	return sel.Controllers(), nil
}

// Instrument

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

// Instruments

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
