package planktoscopes

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "PLANKTOSCOPE_"

type Config struct {
	Planktoscope Planktoscope
}

func GetConfig() (c Config, err error) {
	c.Planktoscope, err = GetPlanktoscope()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Planktoscope config")
	}
	return c, nil
}

func GetPlanktoscope() (p Planktoscope, err error) {
	url, err := env.GetURL(envPrefix+"MJPEGSTREAM", "")
	if err != nil {
		return Planktoscope{}, errors.Wrap(err, "couldn't make server url config")
	}
	p.MJPEGStream = url.String()
	if len(p.MJPEGStream) == 0 {
		return Planktoscope{}, nil
	}

	p.Name = env.GetString(envPrefix+"NAME", url.Host)
	p.Description = env.GetString(
		envPrefix+"DESC",
		"The default Planktoscope device specified in the environment variables.",
	)
	return p, nil
}
