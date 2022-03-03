package conf

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

type HTTPConfig struct {
	GzipLevel int
}

func getHTTPConfig() (c HTTPConfig, err error) {
	const defaultGzipLevel = 1
	rawGzipLevel, err := env.GetInt64("FLUITANS_HTTP_GZIPLEVEL", defaultGzipLevel)
	if err != nil {
		err = errors.Wrap(err, "couldn't make gzip level config")
		return
	}
	c.GzipLevel = int(rawGzipLevel)

	return
}
