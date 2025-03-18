// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

// Payload represents the payload expected by tasks which report orphan
// resources to the Open Delivery Gear API.
type Payload struct {
	// Query represents the SQL query to use when fetching orphan resources.
	Query string `yaml:"query" json:"query"`
}
