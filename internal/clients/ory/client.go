// Package ory provides a high-level client for using Ory Kratos
package ory

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
)

type Client struct {
	Config Config
	Ory    *ory.APIClient
}

func NewClient(c Config) *Client {
	return &Client{
		Config: c,
		Ory:    ory.NewAPIClient(c.KratosAPI),
	}
}

func (c *Client) GetPath(ctx context.Context, endpoint, route string) (string, error) {
	basePath, err := c.Config.KratosAPI.ServerURLWithContext(ctx, endpoint)
	return basePath + route, errors.Wrap(err, "couldn't look up base path for Ory API")
}
