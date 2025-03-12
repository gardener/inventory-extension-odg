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
	// TaskReportOrphanVirtualMachinesAzure is the name of the task, which
	// reports orphan Azure Virtual Machines as findings.
	TaskReportOrphanVirtualMachinesAzure = "odg:task:report-orphan-vms-az"
)

// NewTaskReportOrphanVirtualMachinesAzure creates a new [asynq.Task] for
// reporting orphan Azure Virtual Machines as findings.
func NewTaskReportOrphanVirtualMachinesAzure() *asynq.Task {
	return asynq.NewTask(TaskReportOrphanVirtualMachinesAzure, nil)
}

// HandleReportOrphanVirtualMachinesAzure is a handler, which reports orphan
// Azure virtual machines as findings.
func HandleReportOrphanVirtualMachinesAzure(ctx context.Context, t *asynq.Task) error {
	// TODO: implement me
	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanVirtualMachinesAzure,
		asynq.HandlerFunc(HandleReportOrphanVirtualMachinesAzure),
	)
}
