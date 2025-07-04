// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"
	"time"

	"cloud.google.com/go/civil"
	dbclient "github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/core/registry"
	"github.com/gardener/inventory/pkg/metrics"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"

	apitypes "github.com/gardener/inventory-extension-odg/pkg/odg/api/types"
	odgclient "github.com/gardener/inventory-extension-odg/pkg/odg/client"
	"github.com/gardener/inventory-extension-odg/pkg/odg/models"
)

// TaskReportOrphanVirtualMachinesAWS is the name of the task, which
// reports orphan AWS EC2 Instances as findings.
const TaskReportOrphanVirtualMachinesAWS = "odg:task:report-orphan-vms-aws"

// HandleReportOrphanVirtualMachinesAWS is a handler, which reports orphan AWS
// virtual machines as findings.
func HandleReportOrphanVirtualMachinesAWS(ctx context.Context, t *asynq.Task) error {
	payload, err := DecodePayload(t)
	if err != nil {
		return asynqutils.SkipRetry(err)
	}

	// 1. Fetch orphan resources and create findings out of them
	var items []models.OrphanVirtualMachineAWS
	if err := FetchResourcesFromDB(ctx, dbclient.DB, payload.Query, &items); err != nil {
		return err
	}

	logger := asynqutils.GetLogger(ctx)
	logger.Info("found orphan aws instances", "count", len(items))

	// Metric about discovered orphan resources from Inventory
	metrics.DefaultCollector.AddMetric(
		metrics.Key(TaskReportOrphanVirtualMachinesAWS, "discovered_resources"),
		prometheus.MustNewConstMetric(
			discoveredOrphanResourcesDesc,
			prometheus.GaugeValue,
			float64(len(items)),
			string(apitypes.ProviderNameAWS),
			string(apitypes.ResourceKindVirtualMachineAWS),
		),
	)

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
					ArtefactName:    item.InstanceID,
					ArtefactType:    string(apitypes.ResourceKindVirtualMachineAWS),
					ArtefactVersion: payload.ComponentVersion,
					ArtefactExtraID: map[string]string{
						"vpc_id":      item.VpcID,
						"region_name": item.RegionName,
						"account_id":  item.AccountID,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			Data: apitypes.Finding{
				Severity:     apitypes.SeverityLevelHigh,
				ProviderName: apitypes.ProviderNameAWS,
				ResourceKind: apitypes.ResourceKindVirtualMachineAWS,
				ResourceName: item.InstanceID,
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
					ArtefactName:    item.InstanceID,
					ArtefactType:    string(apitypes.ResourceKindVirtualMachineAWS),
					ArtefactVersion: payload.ComponentVersion,
					ArtefactExtraID: map[string]string{
						"vpc_id":      item.VpcID,
						"region_name": item.RegionName,
						"account_id":  item.AccountID,
					},
				},
				ArtefactKind: apitypes.ArtefactKindRuntime,
			},
			DiscoveryDate: civil.DateOf(now),
		}
		artefacts = append(artefacts, artefact, scanInfo)

		// Runtime artefact item for each finding
		runtimeArtefact := apitypes.ComponentArtefactID{
			ComponentName:    payload.ComponentName,
			ComponentVersion: payload.ComponentVersion,
			Artefact: apitypes.LocalArtefactID{
				ArtefactName:    item.InstanceID,
				ArtefactType:    string(apitypes.ResourceKindVirtualMachineAWS),
				ArtefactVersion: payload.ComponentVersion,
				ArtefactExtraID: map[string]string{
					"vpc_id":      item.VpcID,
					"region_name": item.RegionName,
					"account_id":  item.AccountID,
				},
			},
			ArtefactKind: apitypes.ArtefactKindRuntime,
		}
		runtimeArtefacts = append(runtimeArtefacts, runtimeArtefact)
	}

	// 2. Wipe out old/previous findings for the artefact type.
	oldEntries, err := odgclient.Client.QueryArtefactMetadata(
		ctx,
		apitypes.DatatypeInventory,
		apitypes.ComponentArtefactID{
			ComponentName:    payload.ComponentName,
			ComponentVersion: payload.ComponentVersion,
			ArtefactKind:     apitypes.ArtefactKindRuntime,
			Artefact: apitypes.LocalArtefactID{
				ArtefactType: string(apitypes.ResourceKindVirtualMachineAWS),
			},
		},
	)
	if err != nil {
		return MaybeSkipRetry(err)
	}

	logger.Info("deleting old orphan aws instances from odg", "count", len(oldEntries))
	if err := odgclient.Client.DeleteArtefactMetadata(ctx, oldEntries...); err != nil {
		return MaybeSkipRetry(err)
	}

	// ... also wipe out old runtime artefacts
	labels := map[string]string{
		"created-by":     string(apitypes.DatasourceInventory),
		"resource-kind":  string(apitypes.ResourceKindVirtualMachineAWS),
		"component-name": payload.ComponentName,
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
		"submitting orphan aws instances to odg",
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

	// Metric about successfully reported orphan resources to ODG.
	metrics.DefaultCollector.AddMetric(
		metrics.Key(TaskReportOrphanVirtualMachinesAWS, "reported_resources"),
		prometheus.MustNewConstMetric(
			reportedOrphanResourcesDesc,
			prometheus.GaugeValue,
			float64(len(items)),
			string(apitypes.ProviderNameAWS),
			string(apitypes.ResourceKindVirtualMachineAWS),
		),
	)

	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanVirtualMachinesAWS,
		asynq.HandlerFunc(HandleReportOrphanVirtualMachinesAWS),
	)
}
