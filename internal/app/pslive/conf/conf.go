// Package conf supports environment variable-based application configuration
package conf

import (
	"github.com/pkg/errors"
)

type Config struct {
	DomainName string
	HTTP       HTTPConfig
}

func GetConfig() (c Config, err error) {
	c.HTTP, err = getHTTPConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't make http config")
		return
	}

	return
}
