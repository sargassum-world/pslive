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
	url, err := env.GetURL(envPrefix+"MJPEGSTREAM", "")
	if err != nil {
		return Instrument{}, errors.Wrap(err, "couldn't make server url config")
	}
	p.MJPEGStream = url.String()
	if len(p.MJPEGStream) == 0 {
		return Instrument{}, nil
	}

	p.Name = env.GetString(envPrefix+"NAME", url.Host)
	p.Description = env.GetString(
		envPrefix+"DESC",
		"The default instrument specified in the environment variables.",
	)
	return p, nil
}
