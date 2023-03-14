package instruments

import (
	"time"

	"github.com/pkg/errors"
)

// Job Specification

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
	Controller string `hcl:"controller"`
	Command    string `hcl:"command"`
}

type ParsedSpecification struct {
	Schedule Schedule `hcl:"schedule,block"`
	Action   Action   `hcl:"action,block"`
}
