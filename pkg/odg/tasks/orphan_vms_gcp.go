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

	apitypes "github.com/gardener/inventory-extension-odg/pkg/odg/api/types"
	odgclient "github.com/gardener/inventory-extension-odg/pkg/odg/client"
	"github.com/gardener/inventory-extension-odg/pkg/odg/models"
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
	runtimeArtefacts := make([]apitypes.ComponentArtefactID, 0)
	for _, item := range items {
		// Finding item
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
						"instance_id": strconv.FormatUint(item.InstanceID, 10),
						"project_id":  item.ProjectID,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			Data: apitypes.Finding{
				Severity:     apitypes.SeverityLevelHigh,
				ProviderName: apitypes.ProviderNameGCP,
				ResourceKind: apitypes.ResourceKindVirtualMachineGCP,
				ResourceName: strconv.FormatUint(item.InstanceID, 10),
				Summary:      "Orphan Virtual Machine",
				Attributes:   item,
			},
			DiscoveryDate: civil.DateOf(now),
		}

		// Scan info item for each finding
		scanInfo := apitypes.ArtefactMetadata{
			Meta: apitypes.Metadata{
				Datasource:   apitypes.DatasourceInventory,
				Type:         apitypes.DatatypeArtefactScanInfo,
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
						"instance_id": strconv.FormatUint(item.InstanceID, 10),
						"project_id":  item.ProjectID,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			DiscoveryDate: civil.DateOf(now),
		}
		artefacts = append(artefacts, artefact, scanInfo)

		// Rutime artefact for each finding
		runtimeArtefact := apitypes.ComponentArtefactID{
			ComponentName:    payload.ComponentName,
			ComponentVersion: payload.ComponentVersion,
			Artefact: apitypes.LocalArtefactID{
				ArtefactName:    item.Name,
				ArtefactType:    string(apitypes.ResourceKindVirtualMachineGCP),
				ArtefactVersion: payload.ComponentVersion,
				ArtefactExtraID: map[string]string{
					"instance_id": strconv.FormatUint(item.InstanceID, 10),
					"project_id":  item.ProjectID,
				},
			},
			ArtefactKind: apitypes.ArtefactKindRuntime,
		}
		runtimeArtefacts = append(runtimeArtefacts, runtimeArtefact)
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

	// ... also wipe out old runtime artefacts
	labels := map[string]string{
		"created-by":    string(apitypes.DatasourceInventory),
		"resource-kind": string(apitypes.ResourceKindVirtualMachineGCP),
	}
	oldRuntimeArtefacts, err := odgclient.Client.QueryRuntimeArtefacts(ctx, labels)
	if err != nil {
		return MaybeSkipRetry(err)
	}

	logger.Info("deleting old orphan runtime artefacts from odg", "count", len(oldRuntimeArtefacts))
	runtimeArtefactNames := make([]string, 0)
	for _, raItem := range oldRuntimeArtefacts {
		runtimeArtefactNames = append(runtimeArtefactNames, raItem.Metadata.Name)
	}
	if err := odgclient.Client.DeleteRuntimeArtefacts(ctx, runtimeArtefactNames...); err != nil {
		return MaybeSkipRetry(err)
	}

	// 3. Submit orphan resources from step 1.
	logger.Info(
		"submitting orphan gcp instances to odg",
		"count", len(items),
		"component_name", payload.ComponentName,
		"component_version", payload.ComponentVersion,
	)
	if err := odgclient.Client.SubmitArtefactMetadata(ctx, artefacts...); err != nil {
		return MaybeSkipRetry(err)
	}

	// 4. Submit runtime artefacts
	logger.Info(
		"submitting runtime artefacts",
		"count", len(runtimeArtefacts),
		"component_name", payload.ComponentName,
		"component_version", payload.ComponentVersion,
	)

	if err := odgclient.Client.SubmitRuntimeArtefact(ctx, labels, runtimeArtefacts...); err != nil {
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
