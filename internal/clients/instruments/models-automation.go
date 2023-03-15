package instruments

import (
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
)

// Job Specification

type ParsedSpecification struct {
	Schedule Schedule `hcl:"schedule,block"`
	// TODO: allow non-controller actions
	Actions []Action `hcl:"action,block"`
}

type Schedule struct {
	Interval string `hcl:"interval"`       // a string that parses with time.ParseDuration()
	Start    string `hcl:"start,optional"` // An RFC3339 timestamp
}

func (s Schedule) DecodeStart() (*time.Time, error) {
	// TODO: use cty to handle the decoding instead
	if s.Start == "" {
		return nil, nil
	}

	start, err := time.Parse(time.RFC3339, s.Start)
	return &start, errors.Wrapf(err, "couldn't decode start time %s as rfc3339 timestamp", s.Start)
}

type Action struct {
	Type   string   `hcl:"type,label"`
	Name   string   `hcl:"name,label"`
	Remain hcl.Body `hcl:",remain"`
}

type SleepAction struct {
	Duration string `hcl:"duration"`
}

type ControllerAction struct {
	Controller string   `hcl:"controller"`
	Command    string   `hcl:"command"`
	Params     hcl.Body `hcl:",remain"`
}

type PlanktoscopePumpParams struct {
	Forward  bool    `hcl:"forward"`
	Volume   float64 `hcl:"volume"`
	Flowrate float64 `hcl:"flowrate"`
}
