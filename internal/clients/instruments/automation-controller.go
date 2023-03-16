package instruments

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"
)

type ControllerActionRunner interface {
	RunControllerAction(ctx context.Context, command string, params hcl.Body) error
}

type ControllerActionRunnerGetter func(id ControllerID) (a ControllerActionRunner, ok bool)

// Controller Action Runner Store

type ControllerActionRunnerStore struct {
	instruments     *Store
	protocolGetters map[string]ControllerActionRunnerGetter
}

func NewControllerActionRunnerStore(
	instruments *Store, protocolGetters map[string]ControllerActionRunnerGetter,
) *ControllerActionRunnerStore {
	return &ControllerActionRunnerStore{
		instruments:     instruments,
		protocolGetters: protocolGetters,
	}
}

func (s *ControllerActionRunnerStore) GetActionRunner(
	ctx context.Context, iid InstrumentID, controllerName string,
) (runners map[ControllerID]ControllerActionRunner, err error) {
	controllers, err := s.instruments.GetInstrumentControllersByName(ctx, iid, controllerName)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't lookup controllers named %s for instrument %d", controllerName, iid,
		)
	}
	runners = make(map[ControllerID]ControllerActionRunner)
	for _, controller := range controllers {
		cid := controller.ID
		protocol := controller.Protocol
		getter, ok := s.protocolGetters[protocol]
		if !ok {
			return nil, errors.Errorf("controller %d has unknown protocol %s", cid, protocol)
		}
		if runners[cid], ok = getter(cid); !ok {
			return nil, errors.Errorf("couldn't get action runner for controller %d", cid)
		}
	}
	return runners, nil
}

func (s *ControllerActionRunnerStore) HandleControllerAction(
	ctx context.Context, iid InstrumentID, name string, params hcl.Body,
) error {
	var a ControllerAction
	if err := gohcl.DecodeBody(params, nil, &a); err != nil {
		return errors.Wrapf(err, "couldn't decode controller action %s", name)
	}

	runners, err := s.GetActionRunner(ctx, iid, a.Controller)
	if err != nil {
		return errors.Wrapf(
			err, "couldn't get action runners for controllers named %s on instrument %d",
			a.Controller, iid,
		)
	}
	if len(runners) == 0 {
		return errors.Errorf("couldn't find any controllers named %s", a.Controller)
	}
	for id, runner := range runners {
		if err := runner.RunControllerAction(ctx, a.Command, a.Params); err != nil {
			return errors.Wrapf(err, "couldn't run action %s with controller %d", name, id)
		}
	}
	return nil
}
