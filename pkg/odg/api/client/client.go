// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package client provides an API client for interfacing with the Open Delivery Gear API service.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	apitypes "github.com/gardener/inventory-extension-odg/pkg/odg/api/types"
)

// AuthCookie is the name of the cookie returned by the Delivery
// Service API upon successful authentication.
//
// The cookie must be set on each subsequent request to the API, in order for
// the API calls to be authenticated.
const AuthCookie = "bearer_token"

// ErrNoAuthCookie is an error, which is returned when the remote API server did
// not return an authentication cookie upon successful authentication.
var ErrNoAuthCookie = errors.New("no authentication cookie returned")

// ErrNoGithubAPIURL is an error, which is returned when the [Client] is
// attempting to authenticate, but no Github API URL has been configured.
var ErrNoGithubAPIURL = errors.New("no github api url configured")

// ErrNoGithubToken is an error, which is returned when the [Client] is
// attempting to authenticate, but no Github token has been configured.
var ErrNoGithubToken = errors.New("no github token configured")

// APIError represents an error returned by the remote Delivery Service
// API.
type APIError struct {
	// Method specifies the HTTP method that was used as part of the request
	Method string

	// URL specifies the URL that was used as part of the request
	URL string

	// StatusCode is the HTTP status code returned by the API.
	StatusCode int

	// Body is the body returned as part of the response by the API.
	Body []byte
}

// APIErrorFromResponse creates a new [APIError] from the given [http.Response].
func APIErrorFromResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	apiErr := &APIError{
		Method:     resp.Request.Method,
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	// Add body back to response for future reading
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return apiErr

}

// Error implements the error interface
func (ae *APIError) Error() string {
	s := fmt.Sprintf("method=%s url=%s code=%d body=%s", ae.Method, ae.URL, ae.StatusCode, string(ae.Body))

	return s
}

// Option is a function which configures the [Client].
type Option func(c *Client) error

// Client is an API client for interfacing with the Open Delivery Gear API
// service.
type Client struct {
	// endpoint specifies the remote Delivery Service base API endpoint
	endpoint *url.URL

	// httpClient is the [http.Client], which will be used for API calls
	httpClient *http.Client

	// userAgent specifies the User-Agent header to set when making API
	// calls.
	userAgent string

	// authGithubURL specifies a Github API URL, which the Delivery Service
	// will query for user's information, before signing a JWT token for us.
	authGithubURL *url.URL

	// authGithubToken specifies a Github Personal Access Token (PAT), which
	// the Delivery Service will use to query user's information via the
	// Github API. The information will then be used to create a JWT token,
	// signed with the Delivery Service private keys.
	authGithubToken string
}

// New creates a new [Client] against the provided endpoint and configures it
// using the specified options.
func New(endpoint string, opts ...Option) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{
		endpoint: u,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Configure default HTTP client, unless already set
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	// Make sure that we've got a cookie jar, so that we can store and
	// re-use the authentication cookie.
	if c.httpClient.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		c.httpClient.Jar = jar
	}

	return c, nil
}

// setReqHeaders configures HTTP headers for the given [http.Request]
func (c *Client) setReqHeaders(req *http.Request) {
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

// Authenticate authenticates the API client against the remote Delivery Service
// API.
//
// Upon successful authentication the Delivery Service returns a cookie with a
// JWT bearer token, which will be used in subsequent API calls to the service.
func (c *Client) Authenticate(ctx context.Context) error {
	if c.authGithubURL == nil {
		return ErrNoGithubAPIURL
	}

	if c.authGithubToken == "" {
		return ErrNoGithubToken
	}

	u, err := url.JoinPath(c.endpoint.String(), "/auth")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	c.setReqHeaders(req)
	query := req.URL.Query()
	query.Add("api_url", c.authGithubURL.String())
	query.Add("access_token", c.authGithubToken)
	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return APIErrorFromResponse(resp)
	}

	// Make sure that the API returned our authentication cookie
	gotAuthCookie := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == AuthCookie {
			gotAuthCookie = true
			break
		}
	}

	if !gotAuthCookie {
		return ErrNoAuthCookie
	}

	return nil
}

// Logout logs out from the remote API.
//
// This operation essentially deletes the [AuthCookie] from the cookie jar.
func (c *Client) Logout(ctx context.Context) error {
	u, err := url.JoinPath(c.endpoint.String(), "/auth/logout")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	c.setReqHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return APIErrorFromResponse(resp)
	}

	return nil
}

// QueryArtefactMetadata queries the Delivery Service API for the artefacts of
// the given datatype and described by the specified
// [apitypes.ComponentArtefactID] items.
func (c *Client) QueryArtefactMetadata(
	ctx context.Context,
	datatype apitypes.Datatype,
	items ...apitypes.ComponentArtefactID) ([]apitypes.ArtefactMetadata, error) {
	if len(items) == 0 {
		return nil, nil
	}

	u, err := url.JoinPath(c.endpoint.String(), "/artefacts/metadata/query")
	if err != nil {
		return nil, err
	}

	// Prepare payload for querying artefacts
	payload := apitypes.ComponentArtefactIDGroup{
		Entries: items,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setReqHeaders(req)
	query := req.URL.Query()
	query.Add("type", string(datatype))
	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, APIErrorFromResponse(resp)
	}

	// Parse response body and return results to caller
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []apitypes.ArtefactMetadata
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteArtefactMetadata deletes the given list of [apitypes.ArtefactMetadata]
// from the Delivery Service database.
func (c *Client) DeleteArtefactMetadata(ctx context.Context, items ...apitypes.ArtefactMetadata) error {
	if len(items) == 0 {
		return nil
	}

	u, err := url.JoinPath(c.endpoint.String(), "/artefacts/metadata")
	if err != nil {
		return err
	}

	payload := apitypes.ArtefactMetadataGroup{
		Entries: items,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setReqHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return APIErrorFromResponse(resp)
	}

	return nil
}

// SubmitArtefactMetadata submits the given [apitypes.ArtefactMetadata] items to
// the Delivery Service API.
//
// The provided artefacts are either created, if they don't already exist, or
// are updated when they are already present in the Delivery Service database.
func (c *Client) SubmitArtefactMetadata(ctx context.Context, items ...apitypes.ArtefactMetadata) error {
	if len(items) == 0 {
		return nil
	}

	u, err := url.JoinPath(c.endpoint.String(), "/artefacts/metadata")
	if err != nil {
		return err
	}

	payload := apitypes.ArtefactMetadataGroup{
		Entries: items,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setReqHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return APIErrorFromResponse(resp)
	}

	return nil
}

// QueryRuntimeArtefacts fetches the runtime artefacts with the specified labels
// from the Delivery Service API.
func (c *Client) QueryRuntimeArtefacts(ctx context.Context, labels map[string]string) ([]apitypes.RuntimeArtefactResultItem, error) {
	u, err := url.JoinPath(c.endpoint.String(), "/service-extensions/runtime-artefacts")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setReqHeaders(req)

	// Filter runtime artefacts by label, if specified.
	query := req.URL.Query()
	for k, v := range labels {
		query.Add("label", fmt.Sprintf("%s:%s", k, v))
	}
	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, APIErrorFromResponse(resp)
	}

	// Parse result runtime artefacts
	var result []apitypes.RuntimeArtefactResultItem
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRuntimeArtefacts deletes the runtime artefacts with the specified names
// from the Delivery Service API.
func (c *Client) DeleteRuntimeArtefacts(ctx context.Context, names ...string) error {
	u, err := url.JoinPath(c.endpoint.String(), "/service-extensions/runtime-artefacts")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	c.setReqHeaders(req)

	query := req.URL.Query()
	for _, name := range names {
		query.Add("name", name)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return APIErrorFromResponse(resp)
	}

	return nil
}

// SubmitRuntimeArtefact submits the given [apitypes.ComponentArtefactID] items
// to the Delivery Service API as runtime artefacts.
func (c *Client) SubmitRuntimeArtefact(ctx context.Context, items ...apitypes.ComponentArtefactID) error {
	if len(items) == 0 {
		return nil
	}

	u, err := url.JoinPath(c.endpoint.String(), "/service-extensions/runtime-artefacts")
	if err != nil {
		return err
	}

	payload := apitypes.RuntimeArtefactGroup{
		Artefacts: items,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setReqHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return APIErrorFromResponse(resp)
	}

	return nil
}

// WithGithubAuthentication configures the [Client] to authenticate against the
// remote Delivery Service using a Github access token.
//
// In this authentication mode the Delivery Service queries the user associated
// with the provided access token and then signs a JWT token using it's own
// private key, which is then returned back to the client as a cookie.
//
// Subsequent API calls to the Delivery Service are expected to have the JWT
// token already set as a cookie.
func WithGithubAuthentication(apiURL, accessToken string) Option {
	opt := func(c *Client) error {
		u, err := url.Parse(apiURL)
		if err != nil {
			return err
		}

		c.authGithubURL = u
		c.authGithubToken = accessToken

		return nil
	}

	return opt
}

// WithHTTPClient configures the [Client] to use the specified [http.Client] for
// making calls to the Delivery Service API.
func WithHTTPClient(httpClient *http.Client) Option {
	opt := func(c *Client) error {
		c.httpClient = httpClient

		return nil
	}

	return opt
}

// WithUserAgent configures the [Client] to use the specified User-Agent when
// making API calls.
func WithUserAgent(userAgent string) Option {
	opt := func(c *Client) error {
		c.userAgent = userAgent

		return nil
	}

	return opt
}
