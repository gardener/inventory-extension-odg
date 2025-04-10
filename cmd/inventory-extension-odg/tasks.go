// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"slices"

	"github.com/gardener/inventory/pkg/core/registry"
	"github.com/hibiken/asynq"
	"github.com/urfave/cli/v2"

	_ "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/tasks"
)

// NewTasksCommand returns a new [cli.Command] for tasks-related operations.
func NewTasksCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "task",
		Usage:   "task operations",
		Aliases: []string{"t"},
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "list registered tasks",
				Aliases: []string{"ls"},
				Action:  execTaskListCommand,
			},
		},
	}

	return cmd
}

// execTaskListCommand lists the tasks from the default registry
func execTaskListCommand(ctx *cli.Context) error {
	tasks := make([]string, 0)
	_ = registry.TaskRegistry.Range(func(name string, _ asynq.Handler) error {
		tasks = append(tasks, name)
		return nil
	})

	slices.Sort(tasks)
	for _, name := range tasks {
		fmt.Println(name)
	}

	return nil
}
