package opa

import (
	"net/url"
)

// Resource

type Resource struct {
	URI       string
	ParsedURI *url.URL
}

func NewResource(uri string) Resource {
	parsed, err := url.ParseRequestURI(uri)
	if err != nil {
		return Resource{
			URI: uri,
		}
	}
	return Resource{
		URI:       uri,
		ParsedURI: parsed,
	}
}

func (r Resource) Map() map[string]interface{} {
	if r.ParsedURI == nil {
		return map[string]interface{}{
			"path": r.URI,
		}
	}
	return map[string]interface{}{
		"uri":   r.ParsedURI.String(),
		"path":  r.ParsedURI.Path,
		"query": r.ParsedURI.Query(),
	}
}

// Operation

type Operation struct {
	Method string
	Params interface{}
}

func NewOperation(method string, params interface{}) Operation {
	return Operation{
		Method: method,
		Params: params,
	}
}

func (o Operation) Map() map[string]interface{} {
	return map[string]interface{}{
		"method": o.Method,
		"params": o.Params,
	}
}

// Subject

type Subject struct {
	Identity      string
	Authenticated bool
	Metadata      interface{}
}

func NewSubject(identity string, authenticated bool) Subject {
	return Subject{
		Identity:      identity,
		Authenticated: authenticated,
	}
}

func (s Subject) Map() map[string]interface{} {
	return map[string]interface{}{
		"identity":      s.Identity,
		"authenticated": s.Authenticated,
		"metadata":      s.Metadata,
	}
}

// Input

type Input struct {
	Resource  Resource
	Operation Operation
	Subject   Subject
	Context   interface{}
}

func (i Input) Map() map[string]interface{} {
	return map[string]interface{}{
		"resource":  i.Resource.Map(),
		"operation": i.Operation.Map(),
		"subject":   i.Subject.Map(),
		"context":   i.Context,
	}
}
