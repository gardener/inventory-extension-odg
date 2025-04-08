// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package client provides an API client for interfacing with the Open Delivery Gear API service.
package client

import "net/http"

// Option is a function which configures the [Client].
type Option func(c *Client)

// Client is an API client for interfacing with the Open Delivery Gear API
// service.
type Client struct {
	// endpoint specifies the remote Delivery Service API endpoint
	endpoint string

	// httpClient is the [http.Client], which will be used for API calls
	httpClient *http.Client
}

// New creates a new [Client] against the provided endpoint and configures it
// using the specified options.
func New(endpoint string, opts ...Option) *Client {
	c := &Client{
		endpoint: endpoint,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}
