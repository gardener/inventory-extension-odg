// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package tasks provides tasks for submitting orphan resources discovered by Inventory as
// findings to the Open Delivery Gear API.
//
// The flow for submitting findings to the Delivery Service API is as follows.
//
// 1. Fetch orphan resources from Inventory
//
// Get the orphan resources from Inventory first, then convert them to
// findings, which the Delivery Service understands.
//
// 2. Wipe out old/previous findings for the artefact type
//
// We need to delete the old/previous findings for the artefact type
// associated with the component name and version. This ensures no old
// entries exist in the database, since the Delivery Service does not
// have a retention mechanism for cleaning up such findings.
//
// Also, we need to delete old/previous runtime artefacts for each finding.
//
// 3. Submit the orphan resources from step 1
//
// The latest orphan resources fetched from step 1 are submitted to the
// Delivery Service API.
//
// 4. Create runtime artefacts, so that findings can be evaluated and compliance
// issues created or updated for them.
//
// If we have no orphan resources to report, then we don't report anything to
// the remote API.
package tasks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"

	apiclient "github.com/gardener/inventory-extension-odg/pkg/odg/api/client"
)

// ErrNoPayload is an error, which is returned by task handlers, which expect a
// payload, but none was provided.
var ErrNoPayload = errors.New("no payload specified")

// ErrNoQuery is an error, which is returned by task handlers, which expect a
// query to be provided as part of the payload, but none was provided.
var ErrNoQuery = errors.New("no query specified")

// ErrNoComponentName is an error, which is returned by task handlers, which
// expect an OCM component name to be specified as part of the payload, but none
// was provided.
var ErrNoComponentName = errors.New("no component name specified")

// Payload represents the payload expected by tasks which report orphan
// resources to the Open Delivery Gear API.
type Payload struct {
	// Query represents the SQL query to use when fetching orphan resources.
	Query string `yaml:"query" json:"query"`

	// ComponentName specifies the name of the OCM component with which to
	// associate the submitted findings.
	ComponentName string `yaml:"component_name" json:"component_name"`

	// ComponentVersion specifies the version of the OCM component with
	// which to associate the submitted findings.
	ComponentVersion string `yaml:"component_version" json:"component_version"`
}

// DecodePayload decodes the payload for the given [asynq.Task].
func DecodePayload(t *asynq.Task) (*Payload, error) {
	data := t.Payload()
	if data == nil {
		return nil, ErrNoPayload
	}

	var payload Payload
	if err := asynqutils.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	if payload.Query == "" {
		return nil, ErrNoQuery
	}

	if payload.ComponentName == "" {
		return nil, ErrNoComponentName
	}

	return &payload, nil
}

// FetchResourcesFromDB fetches the resources from the database using the given
// query into the given dest value.
func FetchResourcesFromDB(ctx context.Context, db *bun.DB, query string, dest any) error {
	return db.NewRaw(query).Scan(ctx, dest)
}

// MaybeSkipRetry wraps known API errors with [asynq.SkipRetry], so that the
// tasks which these errors originate from won't be retried.
func MaybeSkipRetry(err error) error {
	// Skip retry for the following HTTP status codes returned by the remote
	// Delivery Service API.
	skipHTTPCodes := []int{
		http.StatusInternalServerError,
	}

	var apiErr *apiclient.APIError
	if errors.As(err, &apiErr) {
		if slices.Contains(skipHTTPCodes, apiErr.StatusCode) {
			return fmt.Errorf("%w (%w)", err, asynq.SkipRetry)
		}
	}

	return err
}
