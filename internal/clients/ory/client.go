// Package ory provides a high-level client for using Ory Kratos
package ory

import (
	ory "github.com/ory/client-go"
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
