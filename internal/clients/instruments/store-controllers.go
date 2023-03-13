package instruments

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
)

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
