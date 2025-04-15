// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"
	"strconv"
	"time"

	"cloud.google.com/go/civil"
	dbclient "github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/core/registry"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"

	apitypes "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/api/types"
	odgclient "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/client"
	"github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/models"
)

// TaskReportOrphanVirtualMachinesGCP is the name of the task, which
// reports orphan GCP Virtual Machines as findings.
const TaskReportOrphanVirtualMachinesGCP = "odg:task:report-orphan-vms-gcp"

// HandleReportOrphanVirtualMachinesGCP is a handler, which reports orphan
// GCP virtual machines as findings.
func HandleReportOrphanVirtualMachinesGCP(ctx context.Context, t *asynq.Task) error {
	payload, err := DecodePayload(t)
	if err != nil {
		return asynqutils.SkipRetry(err)
	}

	var items []models.OrphanVirtualMachineGCP
	if err := FetchResourcesFromDB(ctx, dbclient.DB, payload.Query, &items); err != nil {
		return err
	}

	logger := asynqutils.GetLogger(ctx)
	logger.Info("found orphan gcp instances", "count", len(items))

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
					ArtefactType:    string(apitypes.ResourceKindVirtualMachineGCP),
					ArtefactVersion: payload.ComponentVersion,
					ArtefactExtraID: map[string]string{
						"instance_id": strconv.Itoa(int(item.InstanceID)),
						"project_id":  item.ProjectID,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			Data: apitypes.Finding{
				Severity:     apitypes.SeverityLevelHigh,
				ProviderName: apitypes.ProviderNameGCP,
				ResourceKind: apitypes.ResourceKindVirtualMachineGCP,
				ResourceName: strconv.Itoa(int(item.InstanceID)),
				Summary:      "Orphan Virtual Machine",
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
				ArtefactType: string(apitypes.ResourceKindVirtualMachineGCP),
			},
		},
	)
	if err != nil {
		return MaybeSkipRetry(err)
	}

	logger.Info("deleting old orphan gcp instances from odg", "count", len(oldEntries))
	if err := odgclient.Client.DeleteArtefactMetadata(ctx, oldEntries...); err != nil {
		return MaybeSkipRetry(err)
	}

	// 3. Submit orphan resources from step 1.
	if len(artefacts) == 0 {
		return nil
	}

	logger.Info("submitting orphan gcp instances to odg", "count", len(artefacts))
	if err := odgclient.Client.SubmitArtefactMetadata(ctx, artefacts...); err != nil {
		return MaybeSkipRetry(err)
	}

	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanVirtualMachinesGCP,
		asynq.HandlerFunc(HandleReportOrphanVirtualMachinesGCP),
	)
}
