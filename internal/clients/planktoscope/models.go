package planktoscope

import (
	"time"
)

type Planktoscope struct {
	Pump           Pump
	PumpSettings   PumpSettings
	CameraSettings CameraSettings
	Imager         Imager
	ImagerSettings ImagerSettings
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

func DefaultPumpSettings() PumpSettings {
	const defaultVolume = 1
	const defaultFlowrate = 0.1
	return PumpSettings{
		Forward:  true,
		Volume:   defaultVolume,
		Flowrate: defaultFlowrate,
	}
}

// Camera

type CameraSettings struct {
	StateKnown           bool
	ISO                  uint64
	ShutterSpeed         uint64
	AutoWhiteBalance     bool
	WhiteBalanceRedGain  float64
	WhiteBalanceBlueGain float64
}

func DefaultCameraSettings() CameraSettings {
	const defaultISO = 100
	const defaultShutterSpeed = 125
	const defaultWhiteBalanceRedGain = 2
	const defaultWhiteBalanceBlueGain = 1.4
	return CameraSettings{
		ISO:                  defaultISO,
		ShutterSpeed:         defaultShutterSpeed,
		AutoWhiteBalance:     true,
		WhiteBalanceRedGain:  defaultWhiteBalanceRedGain,
		WhiteBalanceBlueGain: defaultWhiteBalanceBlueGain,
	}
}

// Imager

type Imager struct {
	StateKnown bool
	Imaging    bool
	Start      time.Time
}

type ImagerSettings struct {
	Forward    bool
	StepVolume float64
	StepDelay  float64
	Steps      uint64
}

func DefaultImagerSettings() ImagerSettings {
	const defaultStepVolume = 0.04
	const defaultStepDelay = 0.5
	const defaultSteps = 100
	return ImagerSettings{
		Forward:    true,
		StepVolume: defaultStepVolume,
		StepDelay:  defaultStepDelay,
		Steps:      defaultSteps,
	}
}
