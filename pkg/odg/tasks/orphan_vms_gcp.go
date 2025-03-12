// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"

	"github.com/gardener/inventory/pkg/core/registry"
	"github.com/hibiken/asynq"
)

const (
	// TaskReportOrphanVirtualMachinesGCP is the name of the task, which
	// reports orphan GCP Virtual Machines as findings.
	TaskReportOrphanVirtualMachinesGCP = "odg:task:report-orphan-vms-gcp"
)

// NewTaskReportOrphanVirtualMachinesGCP creates a new [asynq.Task] for
// reporting orphan GCP Virtual Machines as findings.
func NewTaskReportOrphanVirtualMachinesGCP() *asynq.Task {
	return asynq.NewTask(TaskReportOrphanVirtualMachinesGCP, nil)
}

// HandleReportOrphanVirtualMachinesGCP is a handler, which reports orphan
// GCP virtual machines as findings.
func HandleReportOrphanVirtualMachinesGCP(ctx context.Context, t *asynq.Task) error {
	// TODO: implement me
	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanVirtualMachinesGCP,
		asynq.HandlerFunc(HandleReportOrphanVirtualMachinesGCP),
	)
}
