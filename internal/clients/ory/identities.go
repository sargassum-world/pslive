package ory

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
)

// Traits

func getEmail(identity ory.Identity) (string, error) {
	switch schema := identity.SchemaId; schema {
	default:
		traits, ok := identity.Traits.(map[string]interface{})
		if !ok {
			return "", errors.New("couldn't parse Ory Kratos identity traits")
		}
		email, ok := traits["email"].(string)
		if !ok {
			return "", errors.New("couldn't extract email from Ory Kratos identity traits")
		}
		return email, nil
	case "default":
		return "", errors.Errorf("couldn't interpret Ory Kratos identity schema %s", schema)
	}
}

func getIdentifier(identity ory.Identity) (string, error) {
	switch schema := identity.SchemaId; schema {
	default:
		traits, ok := identity.Traits.(map[string]interface{})
		if !ok {
			return "", errors.New("couldn't parse Ory Kratos identity traits")
		}
		username, ok := traits["username"].(string)
		if !ok {
			return "", errors.New("couldn't extract identifier from Ory Kratos identity traits")
		}
		return username, nil
	case "default":
		return "", errors.Errorf("couldn't interpret Ory Kratos identity schema %s", schema)
	}
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
		return "", errors.Wrapf(err, "couldn't look up identity of %s", id)
	}
	identifier, err := getIdentifier(*identity)
	if err != nil {
		return "", errors.Wrapf(err, "couldn't look up identifier for %s", id)
	}
	if err := c.Cache.SetIdentifierByID(id, identifier, c.Config.NetworkCostWeight); err != nil {
		return "", errors.Wrapf(err, "couldn't cache identifier for %s", id)
	}
	return identifier, nil
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
		return Identity{}, errors.Wrapf(err, "couldn't get identity of %s", id)
	}
	parsed, err := parseIdentity(*identity)
	if err != nil {
		return Identity{}, errors.Wrapf(err, "couldn't parse identity of %s", id)
	}

	if err = c.Cache.SetIdentifierByID(
		id, parsed.Identifier, c.Config.NetworkCostWeight,
	); err != nil {
		c.Logger.Error(errors.Wrapf(
			err, "couldn't cache the Ory Kratos identifier for %s", id,
		))
	}
	return parsed, nil
}

// Identities

func (c *Client) GetIdentities(ctx context.Context) ([]Identity, error) {
	oryIdentities, _, err := c.Ory.V0alpha2Api.AdminListIdentities(ctx).Execute()
	identities := make([]Identity, len(oryIdentities))
	for i, identity := range oryIdentities {
		identities[i], err = parseIdentity(identity)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't parse Ory Kratos identity of %s", identity.Id)
		}
		if err = c.Cache.SetIdentifierByID(
			identity.Id, identities[i].Identifier, c.Config.NetworkCostWeight,
		); err != nil {
			c.Logger.Error(errors.Wrapf(
				err, "couldn't cache the Ory Kratos identifier for %s", identity.Id,
			))
		}
	}
	return identities, err
}
