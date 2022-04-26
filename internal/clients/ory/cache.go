package ory

import (
	"fmt"

	"github.com/sargassum-world/fluitans/pkg/godest/clientcache"
)

type Cache struct {
	Cache clientcache.Cache
}

// /ory/identities/:id/identifier

func keyIdentifierByID(id string) string {
	return fmt.Sprintf("/ory/identities/s:[%s]/identifier", id)
}

func (c *Cache) SetIdentifierByID(
	id string, identifier string, costWeight float32,
) error {
	key := keyIdentifierByID(id)
	return c.Cache.SetEntry(key, identifier, costWeight, -1)
}

func (c *Cache) UnsetIdentifierByID(id string) {
	key := keyIdentifierByID(id)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetIdentifierByID(id string) (string, bool, error) {
	key := keyIdentifierByID(id)
	var value string
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return "", keyExists, err
	}

	return value, true, nil
}
