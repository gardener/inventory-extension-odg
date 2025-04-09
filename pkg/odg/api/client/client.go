// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package client provides an API client for interfacing with the Open Delivery Gear API service.
package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

// APIError represents an error returned by the remote Delivery Delivery Service
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
		return fmt.Errorf("cannot ready response body: %w", err)
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

// Authenticate authenticates the API client against the remote Delivery Service
// API.
//
// Upon successful authentication the Delivery Service returns a cookie with a
// JWT bearer token, which will be used in subsequent API calls to the service.
func (c *Client) Authenticate(ctx context.Context) error {
	u, err := url.JoinPath(c.endpoint.String(), "auth")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return err
	}

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
