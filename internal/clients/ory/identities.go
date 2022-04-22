package ory

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
)

// Traits

const identitySchema = "preset://email"

func getEmail(identity ory.Identity) (string, error) {
	switch schema := identity.SchemaId; schema {
	case identitySchema:
		traits, ok := identity.Traits.(map[string]interface{})
		if !ok {
			return "", errors.New("couldn't parse Ory Kratos identity traits")
		}
		email, ok := traits["email"].(string)
		if !ok {
			return "", errors.New("couldn't extract identifier from Ory Kratos identity traits")
		}
		return email, nil
	case "default":
		fallthrough
	default:
		return "", errors.Errorf("couldn't interpret Ory Kratos identity schema %s", schema)
	}
}

func getIdentifier(identity ory.Identity) (string, error) {
	return getEmail(identity)
}

// Identity

type Identity struct {
	ID         string
	Identifier string
	Email      string
}

func parseIdentity(oryIdentity ory.Identity) (identity Identity, err error) {
	identity.ID = oryIdentity.Id
	identity.Identifier, err = getIdentifier(oryIdentity)
	if err != nil {
		return Identity{}, errors.Wrapf(
			err, "couldn't get the identifier for identity %s", oryIdentity.Id,
		)
	}
	identity.Email, err = getEmail(oryIdentity)
	if err != nil {
		return Identity{}, errors.Wrapf(
			err, "couldn't get the email for identity %s", oryIdentity.Id,
		)
	}
	return identity, nil
}

func (c *Client) getIdentifierFromCache(id string) (string, bool) {
	identifier, cacheHit, err := c.Cache.GetIdentifierByID(id)
	if err != nil && err != context.Canceled && errors.Unwrap(err) != context.Canceled {
		// Log the error but return as a cache miss so we can manually query Ory Kratos
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the Ory Kratos identifier for %s", id,
		))
		return "", false // treat an unparseable cache entry like a cache miss
	}

	return identifier, cacheHit
}

func (c *Client) getIdentifierFromOry(ctx context.Context, id string) (string, error) {
	identity, _, err := c.Ory.V0alpha2Api.AdminGetIdentity(ctx, id).Execute()
	if err != nil {
		return "", err
	}
	return getIdentifier(*identity)
}

func (c *Client) GetIdentifier(ctx context.Context, id string) (string, error) {
	if identifier, cacheHit := c.getIdentifierFromCache(id); cacheHit {
		return identifier, nil // empty identifier indicates nonexistent identifier
	}
	return c.getIdentifierFromOry(ctx, id)
}

func (c *Client) GetIdentity(ctx context.Context, id string) (Identity, error) {
	identity, _, err := c.Ory.V0alpha2Api.AdminGetIdentity(ctx, id).Execute()
	if err != nil {
		return Identity{}, err
	}
	return parseIdentity(*identity)
}

// Identities

func (c *Client) GetIdentities(ctx context.Context) ([]Identity, error) {
	oryIdentities, _, err := c.Ory.V0alpha2Api.AdminListIdentities(ctx).Execute()
	identities := make([]Identity, len(oryIdentities))
	for i, identity := range oryIdentities {
		identities[i], err = parseIdentity(identity)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't parse Ory Kratos identity")
		}
	}
	return identities, err
}
