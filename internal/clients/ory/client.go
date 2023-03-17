// Package ory provides a high-level client for using Ory Kratos
package ory

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/clientcache"
)

type Client struct {
	Config Config
	Ory    *ory.APIClient
	Logger godest.Logger
	Cache  *Cache
}

func NewClient(c Config, cache clientcache.Cache, l godest.Logger) *Client {
	return &Client{
		Config: c,
		Ory:    ory.NewAPIClient(c.KratosAPI),
		Logger: l,
		Cache: &Cache{
			Cache: cache,
		},
	}
}

func (c *Client) GetPath(ctx context.Context, endpoint, route string) (string, error) {
	if c.Config.NoAuth {
		return "", nil
	}
	basePath, err := c.Config.KratosAPI.ServerURLWithContext(ctx, endpoint)
	return basePath + route, errors.Wrap(err, "couldn't look up base path for Ory API")
}
