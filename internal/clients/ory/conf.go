package ory

import (
	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "ORY_"

type Config struct {
	KratosAPI         *ory.Configuration
	NetworkCostWeight float32
}

func GetConfig() (c Config, err error) {
	c.KratosAPI = ory.NewConfiguration()

	serverURL, err := env.GetURL(envPrefix+"KRATOS_SERVER", "")
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Ory API server url config")
	}
	c.KratosAPI.Servers = ory.ServerConfigurations{
		{URL: serverURL.String()},
	}

	accessToken := env.GetString(envPrefix+"ACCESS_TOKEN", "")
	if accessToken != "" {
		c.KratosAPI.AddDefaultHeader("Authorization", "Bearer "+accessToken)
	}

	const defaultNetworkCost = 1.0
	c.NetworkCostWeight, err = env.GetFloat32(envPrefix+"NETWORKCOST", defaultNetworkCost)
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make Ory API network cost config")
	}
	return c, nil
}
