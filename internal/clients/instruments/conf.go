package instruments

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "INSTRUMENT_"

type Config struct {
	Instrument Instrument
}

func GetConfig() (c Config, err error) {
	c.Instrument, err = GetInstrument()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Instrument config")
	}
	return c, nil
}

func GetInstrument() (p Instrument, err error) {
	mjpegStreamURL, err := env.GetURL(envPrefix+"MJPEGSTREAM", "")
	if err != nil {
		return Instrument{}, errors.Wrap(err, "couldn't make MJPEG stream url config")
	}
	p.MJPEGStream = mjpegStreamURL.String()
	if len(p.MJPEGStream) == 0 {
		return Instrument{}, nil
	}

	controllerURL, err := env.GetURL(envPrefix+"CONTROLLER", "")
	if err != nil {
		return Instrument{}, errors.Wrap(err, "couldn't make controller url config")
	}
	p.Controller = controllerURL.String()
	if len(p.MJPEGStream) == 0 {
		return Instrument{}, nil
	}

	p.Name = env.GetString(envPrefix+"NAME", mjpegStreamURL.Host)
	p.Description = env.GetString(
		envPrefix+"DESC",
		"The default instrument specified in the environment variables.",
	)
	return p, nil
}
