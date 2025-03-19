// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"

	dbclient "github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/core/registry"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"

	"github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/models"
)

const (
	// TaskReportOrphanPublicAddressGCP is the name of the task, which
	// reports orphan GCP public IP addresses as findings.
	TaskReportOrphanPublicAddressGCP = "odg:task:report-orphan-ip-address-gcp"
)

// HandleReportOrphanPublicAddressGCP is a handler, which reports orphan GCP
// public IP addresses as findings.
func HandleReportOrphanPublicAddressGCP(ctx context.Context, t *asynq.Task) error {
	payload, err := DecodePayload(t)
	if err != nil {
		return asynqutils.SkipRetry(err)
	}

	var items models.OrphanPublicAddressGCP
	if err := FetchResourcesFromDB(ctx, dbclient.DB, payload.Query, &items); err != nil {
		return err
	}

	// TODO: Submit the findings

	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanPublicAddressGCP,
		asynq.HandlerFunc(HandleReportOrphanPublicAddressGCP),
	)
}
