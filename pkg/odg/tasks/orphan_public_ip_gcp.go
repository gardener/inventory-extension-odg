// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	dbclient "github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/core/registry"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"

	apitypes "github.com/gardener/inventory-extension-odg/pkg/odg/api/types"
	odgclient "github.com/gardener/inventory-extension-odg/pkg/odg/client"
	"github.com/gardener/inventory-extension-odg/pkg/odg/models"
)

// TaskReportOrphanPublicAddressGCP is the name of the task, which
// reports orphan GCP public IP addresses as findings.
const TaskReportOrphanPublicAddressGCP = "odg:task:report-orphan-ip-addresses-gcp"

// HandleReportOrphanPublicAddressGCP is a handler, which reports orphan GCP
// public IP addresses as findings.
func HandleReportOrphanPublicAddressGCP(ctx context.Context, t *asynq.Task) error {
	payload, err := DecodePayload(t)
	if err != nil {
		return asynqutils.SkipRetry(err)
	}

	// 1. Fetch orphan resources and create findings out of them
	var items []models.OrphanPublicAddressGCP
	if err := FetchResourcesFromDB(ctx, dbclient.DB, payload.Query, &items); err != nil {
		return err
	}

	logger := asynqutils.GetLogger(ctx)
	logger.Info("found orphan gcp public addresses", "count", len(items))

	now := time.Now()
	artefacts := make([]apitypes.ArtefactMetadata, 0)
	for _, item := range items {
		artefact := apitypes.ArtefactMetadata{
			Meta: apitypes.Metadata{
				Datasource:   apitypes.DatasourceInventory,
				Type:         apitypes.DatatypeInventory,
				CreationDate: now,
				LastUpdate:   now,
			},
			Artefact: apitypes.ComponentArtefactID{
				ComponentName:    payload.ComponentName,
				ComponentVersion: payload.ComponentVersion,
				Artefact: apitypes.LocalArtefactID{
					ArtefactName:    item.Name,
					ArtefactType:    string(apitypes.ResourceKindIPAddressGCP),
					ArtefactVersion: payload.ComponentVersion,
					ArtefactExtraID: map[string]string{
						"project_id":      item.ProjectID,
						"forwarding_rule": item.Name,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			Data: apitypes.Finding{
				Severity:     apitypes.SeverityLevelHigh,
				ProviderName: apitypes.ProviderNameGCP,
				ResourceKind: apitypes.ResourceKindIPAddressGCP,
				ResourceName: fmt.Sprintf("%s:%s", item.ProjectID, item.Name),
				Summary:      "Orphan Public IP Address",
				Attributes:   item,
			},
			DiscoveryDate: civil.DateOf(now),
		}
		artefacts = append(artefacts, artefact)
	}

	// 2. Wipe out old/previous findings for the artefact type
	oldEntries, err := odgclient.Client.QueryArtefactMetadata(
		ctx,
		apitypes.DatatypeInventory,
		apitypes.ComponentArtefactID{
			ComponentName:    payload.ComponentName,
			ComponentVersion: payload.ComponentVersion,
			ArtefactKind:     apitypes.ArtefactKindRuntime,
			Artefact: apitypes.LocalArtefactID{
				ArtefactType: string(apitypes.ResourceKindIPAddressGCP),
			},
		},
	)
	if err != nil {
		return MaybeSkipRetry(err)
	}

	logger.Info("deleting old orphan gcp public ip addresses from odg", "count", len(oldEntries))
	if err := odgclient.Client.DeleteArtefactMetadata(ctx, oldEntries...); err != nil {
		return MaybeSkipRetry(err)
	}

	// 3. Submit orphan resources from step 1.
	if len(artefacts) == 0 {
		return nil
	}

	logger.Info(
		"submitting orphan gcp public ip addresses to odg",
		"count", len(artefacts),
		"component_name", payload.ComponentName,
		"component_version", payload.ComponentVersion,
	)
	if err := odgclient.Client.SubmitArtefactMetadata(ctx, artefacts...); err != nil {
		return MaybeSkipRetry(err)
	}

	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanPublicAddressGCP,
		asynq.HandlerFunc(HandleReportOrphanPublicAddressGCP),
	)
}
