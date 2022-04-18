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
	// Make a request, but insert the CSRF token cookies. Adapted from the impelmentations of the
	// github.com/ory/client-go package's V0alpha2ApiService.SubmitSelfServiceLoginFlowExecute and
	// APIClient.prepareRequest methods.

	// Prepare request body
	body := ory.SubmitSelfServiceLoginFlowWithPasswordMethodBodyAsSubmitSelfServiceLoginFlowBody(
		&ory.SubmitSelfServiceLoginFlowWithPasswordMethodBody{
			Method:     "password",
			CsrfToken:  &csrfToken,
			Identifier: identifier,
			Password:   password,
		},
	)
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, err
	}

	// Prepare request URL
	basePath, err := c.Config.KratosAPI.ServerURLWithContext(
		ctx, "V0alpha2ApiService.SubmitSelfServiceLoginFlow",
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't look up base path for Ory API")
	}
	u, err := url.Parse(basePath + "/self-service/login")
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse Ory API URL %s/self-service/login", basePath)
	}
	if host := c.Config.KratosAPI.Host; host != "" {
		u.Host = host
	}
	if scheme := c.Config.KratosAPI.Scheme; scheme != "" {
		u.Scheme = scheme
	}
	query := u.Query()
	query.Add("flow", flow)
	u.RawQuery = query.Encode()

	// Make request
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), buf)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't make POST request for %s", u.String())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "csrf_token_") {
			req.AddCookie(cookie)
		}
	}
	return req, nil
}

func (c *Client) SubmitLoginFlow(
	ctx context.Context, flow, csrfToken, identifier, password string, cookies []*http.Cookie,
) (*ory.SuccessfulSelfServiceLoginWithoutBrowser, error) {
	req, err := c.makeSubmitLoginFlowRequest(ctx, flow, csrfToken, identifier, password, cookies)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make Ory Kratos submit login flow request")
	}
	res, err := c.Config.KratosAPI.HTTPClient.Do(req)
	if err != nil || res == nil {
		return nil, errors.Wrap(err, "couldn't perform Ory Kratos submit login flow request")
	}

	// Process the response. Adapted from the implemtnation of the github.com/ory/client-go package's
	// V0alpha2ApiService.SubmitSelfServiceLoginFlowExecute method
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read Ory Kratos submit login flow response body")
	}
	if err := res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "couldn't close Ory Kratos submit login flow response body")
	}
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if res.StatusCode >= http.StatusMultipleChoices { // i.e. 300
		// TODO: parse and handle the various error codes
		return nil, errors.Errorf("ory api error %d", res.StatusCode)
	}
	jsonBody := &ory.SuccessfulSelfServiceLoginWithoutBrowser{}
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal Ory Kratos submit login flow response body")
	}
	return jsonBody, nil
}
