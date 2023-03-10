package ory

import (
	"fmt"

	"github.com/sargassum-world/godest/clientcache"
)

type Cache struct {
	Cache clientcache.Cache
}

// /ory/identities/:id/identifier

func keyIdentifierByID(id IdentityID) string {
	return fmt.Sprintf("/ory/identities/s:[%s]/identifier", id)
}

func (c *Cache) SetIdentifierByID(
	id IdentityID, identifier IdentityIdentifier, costWeight float32,
) error {
	key := keyIdentifierByID(id)
	return c.Cache.SetEntry(key, identifier, costWeight, -1)
}

func (c *Cache) UnsetIdentifierByID(id IdentityID) {
	key := keyIdentifierByID(id)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetIdentifierByID(id IdentityID) (IdentityIdentifier, bool, error) {
	key := keyIdentifierByID(id)
	var value string
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return "", keyExists, err
	}

	return IdentityIdentifier(value), true, nil
}
