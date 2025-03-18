// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import "errors"

// ErrNoPayload is an error, which is returned by task handlers, which expect a
// payload, but none was specified.
var ErrNoPayload = errors.New("no payload specified")

// ErrNoQuery is an error, which is returned by task handlers, which expect a
// query to be provided as part of the payload, but none was specified.
var ErrNoQuery = errors.New("no query specified")

// Payload represents the payload expected by tasks which report orphan
// resources to the Open Delivery Gear API.
type Payload struct {
	// Query represents the SQL query to use when fetching orphan resources.
	Query string `yaml:"query" json:"query"`
}
