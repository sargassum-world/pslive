// Package instruments provides a high-level client for management of imaging instruments
package instruments

import (
	"github.com/sargassum-world/fluitans/pkg/godest"
)

type Client struct {
	Config Config
	Logger godest.Logger
}

func NewClient(c Config, l godest.Logger) *Client {
	return &Client{
		Config: c,
		Logger: l,
	}
}
