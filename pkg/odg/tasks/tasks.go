// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"
	"errors"

	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"
)

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

	return &payload, nil
}

// FetchResourcesFromDB fetches the resources from the database using the given
// query into the given dest value.
func FetchResourcesFromDB(ctx context.Context, db *bun.DB, query string, dest any) error {
	return db.NewRaw(query).Scan(ctx, dest)
}
