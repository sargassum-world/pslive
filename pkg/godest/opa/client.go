// Package opa provides a high-level client for using embedded Rego policies with OPA
package opa

import (
	"context"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
)

func NewRouteInput(method, path, identity string, authenticated bool) map[string]interface{} {
	return map[string]interface{}{
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
	}
}

type Client struct {
	allowQuery  rego.PreparedPartialQuery
	errorsQuery rego.PartialResult
}

func makeOptions(query string, options []func(r *rego.Rego)) []func(r *rego.Rego) {
	return append(
		[]func(r *rego.Rego){
			rego.Query(query),
		},
		options...,
	)
}

func NewOptimizedQuery(
	ctx context.Context, query string, options ...func(r *rego.Rego),
) (rego.PartialResult, error) {
	optimized, err := rego.New(makeOptions(query, options)...).PartialResult(ctx)
	return optimized, errors.Wrap(
		err, "couldn't perform pre-input policy optimization",
	)
}

func NewPartialEvalQuery(
	ctx context.Context, query string, unknowns []string, options ...func(r *rego.Rego),
) (rego.PreparedPartialQuery, error) {
	optimized, err := NewOptimizedQuery(ctx, query, options...)
	if err != nil {
		return rego.PreparedPartialQuery{}, err
	}

	prepared, err := optimized.Rego(rego.Unknowns(unknowns)).PrepareForPartial(ctx)
	return prepared, errors.Wrap(err, "couldn't prepare for post-input partial evaluation")
}

func NewClient(entryPackage string, options ...func(r *rego.Rego)) (*Client, error) {
	ctx := context.TODO()
	allowQuery, err := NewPartialEvalQuery(
		ctx, entryPackage+".allow", []string{"input.context"}, options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare policy query for allow")
	}
	errorsQuery, err := NewOptimizedQuery(
		ctx, entryPackage+".errors", options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare policy query for errors")
	}
	return &Client{
		allowQuery:  allowQuery,
		errorsQuery: errorsQuery,
	}, nil
}

func (c *Client) EvalAllow(
	ctx context.Context, input map[string]interface{},
) (allow bool, remainingQueries []ast.Body, err error) {
	pq, evalErr := c.allowQuery.Partial(ctx, rego.EvalInput(input))
	if evalErr != nil {
		return false, nil, errors.Wrap(
			evalErr, "couldn't partially evaluate policy with initial inputs for allow",
		)
	}
	if len(pq.Queries) == 0 {
		return false, nil, nil
	}
	remainingQueries = make([]ast.Body, 0, len(pq.Queries))
	for _, query := range pq.Queries {
		if len(query) == 0 {
			return true, nil, nil
		}
		remainingQueries = append(remainingQueries, query)
		// TODO: translate the non-empty rego query into an SQL query
	}
	return false, remainingQueries, nil
}

func (c *Client) EvalErrors(
	ctx context.Context, input map[string]interface{},
) (authzErr error, evalErr error) {
	results, evalErr := c.errorsQuery.Rego(rego.Input(input)).Eval(ctx)
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for errors")
	}
	return parseErrorsResults(results)
}

func parseSingleExpression(results rego.ResultSet) (expression rego.ExpressionValue, err error) {
	if len(results) != 1 {
		return rego.ExpressionValue{}, errors.Errorf("expected one result but got %d", len(results))
	}
	expressions := results[0].Expressions
	if len(expressions) != 1 {
		return rego.ExpressionValue{}, errors.Errorf(
			"expected one expression in result but got %d", len(expressions),
		)
	}
	if expressions[0] == nil {
		return rego.ExpressionValue{}, errors.New("undefined expression in result")
	}
	expression = *(expressions[0])
	return expression, nil
}

func parseErrorsResults(results rego.ResultSet) (authzErr error, parseErr error) {
	errorsExpression, parseErr := parseSingleExpression(results)
	if parseErr != nil {
		return nil, errors.Errorf("result set doesn't have exactly one expression")
	}
	errs, ok := errorsExpression.Value.([]interface{})
	if !ok {
		return nil, errors.Errorf("errors result has unexpected type %T", errorsExpression)
	}
	authzErrs := make([]string, len(errs))
	for i, err := range errs {
		authzErrs[i] = fmt.Sprintf("%s", err)
	}
	return errors.New(strings.Join(authzErrs, "; ")), parseErr
}
