// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package client provides an API client for interfacing with the Open Delivery Gear API service.
package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

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

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
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

	// TODO: Better errors
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
