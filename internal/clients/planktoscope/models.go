package planktoscope

import (
	"time"
)

type Planktoscope struct {
	Pump         Pump
	PumpSettings PumpSettings
}

// Pump

type Pump struct {
	StateKnown bool
	Pumping    bool
	Start      time.Time
	Duration   time.Duration
	Deadline   time.Time
}

type PumpSettings struct {
	Forward  bool
	Volume   float64
	Flowrate float64
}
