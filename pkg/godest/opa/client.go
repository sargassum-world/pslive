// Package opa provides a high-level client for using embedded Rego policies with OPA
package opa

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
)

type Client struct {
	authzQuery rego.PreparedEvalQuery
}

func NewClient(entryPackage string, options ...func(r *rego.Rego)) (*Client, error) {
	options = append(
		[]func(r *rego.Rego){
			rego.Query(fmt.Sprintf(
				"allow = %s.allow; errors = %s.errors", entryPackage, entryPackage,
			)),
		},
		options...,
	)
	authzQuery, err := rego.New(options...).PrepareForEval(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare OPA policy")
	}
	return &Client{
		authzQuery: authzQuery,
	}, nil
}

func (c *Client) CheckRoute(
	ctx context.Context, method, path string, authenticated bool, identity string,
) (allow bool, authzErr error, evalErr error) {
	results, evalErr := c.authzQuery.Eval(ctx, rego.EvalInput(map[string]interface{}{
		"operation": map[string]interface{}{
			"method": method,
		},
		"resource": map[string]interface{}{
			"path": path,
		},
		"subject": map[string]interface{}{
			"authenticated": authenticated,
			"identity":      identity,
		},
	}))
	if evalErr != nil {
		return false, nil, errors.Wrap(evalErr, "couldn't evaluate OPA policy")
	}
	// TODO: perform any necessary SQL lookups
	allow, authzErr = parseResults(results)
	return allow, errors.Wrap(authzErr, "unauthorized"), nil
}

func parseResults(results rego.ResultSet) (bool, error) {
	if len(results) != 1 {
		return false, errors.Errorf("expected one result but got %d", len(results))
	}

	// Check allow
	allowBinding, ok := results[0].Bindings["allow"]
	if !ok {
		return false, errors.Errorf("result set missing allow")
	}
	allow, ok := allowBinding.(bool)
	if !ok {
		return false, errors.Errorf("allow result has unexpected type %T", allowBinding)
	}

	// Check errors
	errorsBinding, ok := results[0].Bindings["errors"]
	if !ok {
		return false, errors.Errorf("result set missing errors")
	}
	e, ok := errorsBinding.([]interface{})
	if !ok {
		return false, errors.Errorf("errors result has unexpected type %T", errorsBinding)
	}
	if len(e) == 0 {
		return allow, errors.New("unknown")
	}
	if len(e) == 1 {
		return false, errors.Errorf("%s", e[0])
	}
	return false, errors.Errorf("%s", e)
}
