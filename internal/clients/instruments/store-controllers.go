package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

// Enabled Controllers by Protocol

//go:embed queries/select-enabled-controllers-by-protocol.sql
var rawSelectEnabledControllersByProtocolQuery string

var selectEnabledControllersByProtocolQuery string = strings.TrimSpace(
	rawSelectEnabledControllersByProtocolQuery,
)

func (s *Store) GetEnabledControllersByProtocol(
	ctx context.Context, protocol string,
) (controllers []Controller, err error) {
	sel := newControllersSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectEnabledControllersByProtocolQuery,
		Controller{Protocol: protocol}.newProtocolSelection(),
		sel.Step,
	); err != nil {
		return nil, errors.Wrapf(err, "couldn't get enabled controllers with protocol %s", protocol)
	}
	return sel.Controllers(), nil
}

// Instrument Controllers by Name

//go:embed queries/select-instrument-controllers-by-name.sql
var rawSelectInstrumentControllersByNameQuery string

var selectInstrumentControllersByNameQuery string = strings.TrimSpace(
	rawSelectInstrumentControllersByNameQuery,
)

func (s *Store) GetInstrumentControllersByName(
	ctx context.Context, instrumentID InstrumentID, name string,
) (controllers []Controller, err error) {
	sel := newControllersSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectInstrumentControllersByNameQuery,
		Controller{
			InstrumentID: instrumentID,
			Name:         name,
		}.newProtocolSelection(),
		sel.Step,
	); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't get instrument %d controllers with name %s", instrumentID, name,
		)
	}
	return sel.Controllers(), nil
}
