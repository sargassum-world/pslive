package ory

import (
	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "ORY_"

type Config struct {
	KratosAPI *ory.Configuration
}

func GetConfig() (c Config, err error) {
	c.KratosAPI = ory.NewConfiguration()

	serverURL, err := env.GetURL(envPrefix+"KRATOS_SERVER", "")
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make server url config")
	}
	c.KratosAPI.Servers = ory.ServerConfigurations{
		{URL: serverURL.String()},
	}

	return c, nil
}
