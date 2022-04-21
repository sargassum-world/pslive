package ory

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	ory "github.com/ory/client-go"
	"github.com/pkg/errors"
)

func filterCookies(cookies []*http.Cookie, prefixes ...string) []*http.Cookie {
	filtered := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		for _, prefix := range prefixes {
			if strings.HasPrefix(cookie.Name, prefix) {
				filtered = append(filtered, cookie)
			}
		}
	}
	return filtered
}

func (c *Client) makeSelfServiceRequest(
	ctx context.Context, method string, endpoint, route string, query url.Values,
	body interface{}, cookies []*http.Cookie,
) (*http.Request, error) {
	// Make a request, but insert the CSRF token cookies. Adapted from the implementations of the
	// github.com/ory/client-go package's V0alpha2ApiService.SubmitSelfServiceLoginFlowExecute and
	// APIClient.prepareRequest methods.
	// Prepare request body
	buf := &bytes.Buffer{}
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, errors.Wrapf(err, "couldn't json-marshal request body %T", body)
		}
	}

	// Prepare request URL
	basePath, err := c.Config.KratosAPI.ServerURLWithContext(ctx, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't look up base path for Ory API")
	}
	u, err := url.Parse(basePath + route)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse Ory API URL %s%s", basePath, route)
	}
	if host := c.Config.KratosAPI.Host; host != "" {
		u.Host = host
	}
	if scheme := c.Config.KratosAPI.Scheme; scheme != "" {
		u.Scheme = scheme
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't make POST request for %s", u.String())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	return req, nil
}

// Login flow

func (c *Client) InitializeLoginFlow(
	ctx context.Context,
) (*ory.SelfServiceLoginFlow, *http.Cookie, error) {
	flow, res, err := c.Ory.V0alpha2Api.InitializeSelfServiceLoginFlowForBrowsers(ctx).Execute()
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't initialize Ory Kratos self-service login flow")
	}
	if err := res.Body.Close(); err != nil {
		return nil, nil, errors.Wrap(
			err, "couldn't close Ory Kratos self-service login flow response body",
		)
	}

	return flow, res.Cookies()[0], nil
}

func (c *Client) makeSubmitLoginFlowRequest(
	ctx context.Context, flow, csrfToken, identifier, password string, cookies []*http.Cookie,
) (*http.Request, error) {
	body := ory.SubmitSelfServiceLoginFlowWithPasswordMethodBodyAsSubmitSelfServiceLoginFlowBody(
		&ory.SubmitSelfServiceLoginFlowWithPasswordMethodBody{
			Method:     "password",
			CsrfToken:  &csrfToken,
			Identifier: identifier,
			Password:   password,
		},
	)
	query := make(url.Values)
	query.Add("flow", flow)
	return c.makeSelfServiceRequest(
		ctx, http.MethodPost, "V0alpha2ApiService.SubmitSelfServiceLoginFlow", "/self-service/login",
		query, body, filterCookies(cookies, "csrf_token_"),
	)
}

func (c *Client) SubmitLoginFlow(
	ctx context.Context, flow, csrfToken, identifier, password string, cookies []*http.Cookie,
) (*ory.SuccessfulSelfServiceLoginWithoutBrowser, []*http.Cookie, error) {
	req, err := c.makeSubmitLoginFlowRequest(ctx, flow, csrfToken, identifier, password, cookies)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't make Ory Kratos submit login flow request")
	}
	res, err := c.Config.KratosAPI.HTTPClient.Do(req)
	if err != nil || res == nil {
		return nil, nil, errors.Wrap(err, "couldn't perform Ory Kratos submit login flow request")
	}

	// Process the response. Adapted from the implementation of the github.com/ory/client-go package's
	// V0alpha2ApiService.SubmitSelfServiceLoginFlowExecute method
	// TODO: move this into a private utility method
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't read Ory Kratos submit login flow response body")
	}
	if err := res.Body.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "couldn't close Ory Kratos submit login flow response body")
	}
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if res.StatusCode >= http.StatusMultipleChoices { // i.e. 300
		// TODO: parse and handle the various error codes
		return nil, nil, errors.Errorf("ory login flow response error %d", res.StatusCode)
	}
	jsonBody := &ory.SuccessfulSelfServiceLoginWithoutBrowser{}
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		return nil, nil, errors.Wrap(err, "couldn't unmarshal Ory Kratos submit login flow response body")
	}
	return jsonBody, res.Cookies(), nil
}

// Logout flow

func (c *Client) makeSubmitLogoutFlowRequest(
	ctx context.Context, token string, cookies []*http.Cookie,
) (*http.Request, error) {
	query := make(url.Values)
	query.Add("token", token)
	return c.makeSelfServiceRequest(
		ctx, http.MethodGet, "V0alpha2ApiService.SubmitSelfServiceLogoutFlow", "/self-service/logout",
		query, nil, filterCookies(cookies, "ory_session_"),
	)
}

func (c *Client) PerformLogout(
	ctx context.Context, cookies []*http.Cookie,
) ([]*http.Cookie, error) {
	// Initialize logout flow
	var merged string
	for _, cookie := range filterCookies(cookies, "ory_session_") {
		merged += cookie.String() + ";"
	}
	url, res, err := c.Ory.V0alpha2Api.CreateSelfServiceLogoutFlowUrlForBrowsers(
		ctx,
	).Cookie(merged).Execute()
	if err != nil || url == nil {
		return res.Cookies(), errors.Wrap(err, "couldn't create Ory Kratos logout url")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(
			err, "couldn't close Ory Kratos self-service logout flow response body",
		)
	}

	// Submit flow
	req, err := c.makeSubmitLogoutFlowRequest(ctx, url.LogoutToken, cookies)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Ory Kratos submit logout flow request")
	}
	res, err = c.Config.KratosAPI.HTTPClient.Do(req)
	if err != nil || res == nil {
		return nil, errors.Wrap(err, "couldn't perform Ory Kratos submit logout flow request")
	}
	if res.StatusCode >= http.StatusMultipleChoices { // i.e. 300
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't read Ory Kratos submit logout flow response body")
		}
		if err := res.Body.Close(); err != nil {
			return nil, errors.Wrap(err, "couldn't close Ory Kratos submit logout flow response body")
		}
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		// TODO: parse and handle the various error codes
		return nil, errors.Errorf(
			"ory logout flow response error %d: %s", res.StatusCode, string(body),
		)
	}
	return res.Cookies(), nil
}
