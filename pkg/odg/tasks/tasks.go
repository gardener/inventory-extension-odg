// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"context"

	"github.com/gardener/inventory/pkg/common/utils"
	"github.com/gardener/inventory/pkg/core/registry"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	"github.com/hibiken/asynq"
)

const (
	// TaskSubmitAllFindings is the name of the meta task, which enqueues
	// tasks for submitting all supported findings to the Open Delivery Gear
	// API service.
	TaskSubmitAllFindings = "odg:task:submit-all-findings"
)

// HandleSubmitAllFindings is a handler, which enqueues tasks for submitting all
// supported findings to the Open Delivery Gear API service.
func HandleSubmitAllFindings(ctx context.Context, t *asynq.Task) error {
	queue := asynqutils.GetQueueName(ctx)
	taskFuncs := []utils.TaskConstructor{}

	return utils.Enqueue(ctx, taskFuncs, asynq.Queue(queue))
}

// init registers the task handlers with the default Inventory registry
func init() {
	registry.TaskRegistry.MustRegister(TaskSubmitAllFindings, asynq.HandlerFunc(HandleSubmitAllFindings))
}
