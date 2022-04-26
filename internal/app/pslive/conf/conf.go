// Package conf supports environment variable-based application configuration
package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type Config struct {
	Cache ristretto.Config
	HTTP  HTTPConfig
}

func GetConfig() (c Config, err error) {
	c.Cache, err = getCacheConfig()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make cache config")
	}

	c.HTTP, err = getHTTPConfig()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make http config")
	}
	return c, nil
}
