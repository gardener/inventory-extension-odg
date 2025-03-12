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
	// TaskReportOrphanVirtualMachinesAWS is the name of the task, which
	// reports orphan AWS EC2 Instances as findings.
	TaskReportOrphanVirtualMachinesAWS = "odg:task:report-orphan-vms-aws"
)

// NewTaskReportOrphanVirtualMachinesAWS creates a new [asynq.Task] for
// reporting orphan AWS EC2 Virtual Machines as findings.
func NewTaskReportOrphanVirtualMachinesAWS() *asynq.Task {
	return asynq.NewTask(TaskReportOrphanVirtualMachinesAWS, nil)
}

// HandleReportOrphanVirtualMachinesAWS is a handler, which reports orphan AWS
// virtual machines as findings.
func HandleReportOrphanVirtualMachinesAWS(ctx context.Context, t *asynq.Task) error {
	// TODO: implement me
	return nil
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(
		TaskReportOrphanVirtualMachinesAWS,
		asynq.HandlerFunc(HandleReportOrphanVirtualMachinesAWS),
	)
}
