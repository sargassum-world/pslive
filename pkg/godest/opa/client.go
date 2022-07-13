// Package opa provides a high-level client for using embedded Rego policies with OPA
package opa

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/pkg/errors"
)

type Client struct {
	allowQuery           rego.PartialResult
	allowContextualQuery rego.PreparedPartialQuery
	errorQuery           rego.PartialResult
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
	if topdown.IsCancel(err) {
		return rego.PartialResult{}, context.Canceled
	}
	return optimized, errors.Wrap(
		err, "couldn't perform pre-input policy optimization",
	)
}

func NewPartialEvalQuery(
	ctx context.Context, query string, unknowns []string, options ...func(r *rego.Rego),
) (rego.PreparedPartialQuery, error) {
	prepared, err := rego.New(makeOptions(
		query,
		append(options, rego.Unknowns(unknowns)),
	)...).PrepareForPartial(ctx)
	if topdown.IsCancel(err) {
		return rego.PreparedPartialQuery{}, context.Canceled
	}
	return prepared, errors.Wrap(err, "couldn't prepare for partial evaluation")
}

func NewClient(entryPackage string, options ...func(r *rego.Rego)) (*Client, error) {
	ctx := context.TODO()

	allowQuery, err := NewOptimizedQuery(
		ctx, entryPackage+".allow", options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare fast policy query for allow")
	}

	allowContextualQuery, err := NewPartialEvalQuery(
		ctx, entryPackage+".allow", []string{"input.context"}, options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare contextual policy query for allow")
	}
	errorQuery, err := NewOptimizedQuery(
		ctx, entryPackage+".error", options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't prepare policy query for errors")
	}
	return &Client{
		allowQuery:           allowQuery,
		allowContextualQuery: allowContextualQuery,
		errorQuery:           errorQuery,
	}, nil
}

func (c *Client) EvalAllow(
	ctx context.Context, input map[string]interface{},
) (allow bool, err error) {
	results, evalErr := c.allowQuery.Rego(rego.Input(input)).Eval(ctx)
	if topdown.IsCancel(evalErr) {
		return false, context.Canceled
	}
	if evalErr != nil {
		return false, errors.Wrap(evalErr, "couldn't evaluate policy for allow")
	}
	return parseAllowResults(results)
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

func parseAllowResults(results rego.ResultSet) (allow bool, parseErr error) {
	if len(results) == 0 {
		return false, nil
	}
	errorExpression, parseErr := parseSingleExpression(results)
	if parseErr != nil {
		return false, errors.Wrap(parseErr, "allow result set has more than one expression")
	}
	resultAllow, ok := errorExpression.Value.(bool)
	if !ok {
		return false, errors.Errorf("error result has unexpected type %T", errorExpression.Value)
	}
	return resultAllow, parseErr
}

func (c *Client) EvalAllowContextual(
	ctx context.Context, input map[string]interface{},
) (allow bool, remainingQueries []ast.Body, err error) {
	// Note: this function call is very slow (by factor of ~10) when Go's data race detector is active
	pq, evalErr := c.allowContextualQuery.Partial(ctx, rego.EvalInput(input))
	if topdown.IsCancel(evalErr) {
		return false, nil, context.Canceled
	}
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

func (c *Client) EvalError(
	ctx context.Context, input map[string]interface{},
) (authzErr error, evalErr error) {
	results, evalErr := c.errorQuery.Rego(rego.Input(input)).Eval(ctx)
	if topdown.IsCancel(evalErr) {
		return nil, context.Canceled
	}
	if evalErr != nil {
		return nil, errors.Wrap(evalErr, "couldn't evaluate policy for errors")
	}
	return parseErrorResults(results)
}

func parseErrorResults(results rego.ResultSet) (authzErr error, parseErr error) {
	if len(results) == 0 {
		return nil, nil // no policy-reported errors to retrieve
	}
	errorExpression, parseErr := parseSingleExpression(results)
	if parseErr != nil {
		return nil, errors.Wrap(parseErr, "error result set has more than one expression")
	}
	resultError, ok := errorExpression.Value.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("error result has unexpected type %T", errorExpression.Value)
	}
	message, ok := resultError["message"]
	if !ok {
		return nil, errors.Errorf("error result has no message")
	}
	return errors.New(fmt.Sprintf("%s", message)), parseErr
}
