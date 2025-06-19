// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/gardener/inventory/pkg/metrics"
)

var (
	// discoveredOrphanResourcesDesc is the descriptor for a metric, which
	// tracks the number of discovered orphan resources when querying
	// Inventory.
	discoveredOrphanResourcesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, "", "odg_discovered_orphan_resources"),
		"A gauge which tracks the number of discovered orphan resources from Inventory",
		[]string{"provider_name", "resource_kind"},
		nil,
	)

	// reportedOrphanResourcesDesc is the descriptor for a metric, which
	// tracks the number of successfully reported orphan resources to the
	// Open Delivery Gear API.
	reportedOrphanResourcesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, "", "odg_reported_orphan_resources"),
		"A gauge which tracks the number of successfully reported orphan resources to ODG",
		[]string{"provider_name", "resource_kind"},
		nil,
	)
)

// init registers the metric descriptors with [metrics.DefaultCollector]
func init() {
	metrics.DefaultCollector.AddDesc(
		discoveredOrphanResourcesDesc,
		reportedOrphanResourcesDesc,
	)
}
