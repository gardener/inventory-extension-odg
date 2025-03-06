// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"

	dbclient "github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/core/registry"
	asynqutils "github.com/gardener/inventory/pkg/utils/asynq"
	workerutils "github.com/gardener/inventory/pkg/utils/asynq/worker"
	dbutils "github.com/gardener/inventory/pkg/utils/db"
	slogutils "github.com/gardener/inventory/pkg/utils/slog"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/urfave/cli/v2"

	"github.tools.sap/kubernetes/inventory-extension-odg/pkg/config"
	"github.tools.sap/kubernetes/inventory-extension-odg/pkg/version"
)

func main() {
	app := &cli.App{
		Name:                 "inventory-extension-odg",
		Version:              version.Version,
		EnableBashCompletion: true,
		Usage:                "inventory extension for open delivery gear",
		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "start worker process",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "config",
						Usage:    "path to extension config file",
						Required: true,
						Aliases:  []string{"file"},
						EnvVars:  []string{"INVENTORY_EXTENSION_CONFIG"},
					},
				},
				Action: execStartCommand,
			},
			{
				Name:    "tasks",
				Usage:   "list registered tasks",
				Aliases: []string{"t"},
				Action:  execListTasksCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// newDB creates a new [bun.DB] database client based on the given config.
func newDB(conf *config.Config) (*bun.DB, error) {
	db, err := dbutils.NewFromConfig(conf.Database)
	if err != nil {
		return nil, err
	}
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(conf.Debug)))

	return db, nil
}

// newWorker creates a new [workerutils.Worker] from the given config.
func newWorker(conf *config.Config) *workerutils.Worker {
	redisClientOpt := asynqutils.NewRedisClientOptFromConfig(conf.Redis)

	opts := make([]workerutils.Option, 0)
	logLevel := asynq.InfoLevel
	if conf.Debug {
		logLevel = asynq.DebugLevel
	}

	opts = append(opts, workerutils.WithLogLevel(logLevel))
	opts = append(opts, workerutils.WithErrorHandler(asynqutils.NewDefaultErrorHandler()))
	worker := workerutils.NewFromConfig(redisClientOpt, conf.Worker, opts...)

	// Configure middlewares
	middlewares := []asynq.MiddlewareFunc{
		asynqutils.NewLoggerMiddleware(slog.Default()),
		asynqutils.NewMeasuringMiddleware(),
	}
	worker.UseMiddlewares(middlewares...)

	return worker
}

// execStartCommand starts the worker
func execStartCommand(ctx *cli.Context) error {
	// Parse config files for the extension
	configPaths := ctx.StringSlice("config")
	conf, err := config.Parse(configPaths...)
	if err != nil {
		return fmt.Errorf("Cannot parse config: %w", err)
	}

	// Configure the default [slog.Logger]
	logger, err := slogutils.NewFromConfig(os.Stdout, conf.Logging)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	// Configure database client and set it up for task handlers
	db, err := newDB(conf)
	if err != nil {
		return err
	}
	dbclient.SetDB(db)
	defer db.Close()

	// Create a worker, register handlers and start it up
	worker := newWorker(conf)
	worker.HandlersFromRegistry(registry.TaskRegistry)
	registry.TaskRegistry.Range(func(name string, _ asynq.Handler) error {
		slog.Info("registered task", "name", name)
		return nil
	})
	slog.Info("worker concurrency", "level", conf.Worker.Concurrency)
	slog.Info("queue priority", "strict", conf.Worker.StrictPriority)
	for queue, priority := range conf.Worker.Queues {
		slog.Info("queue configuration", "name", queue, "priority", priority)
	}

	return worker.Run()
}

// execListTasksCommand lists the tasks from the default registry
func execListTasksCommand(ctx *cli.Context) error {
	tasks := make([]string, 0)
	registry.TaskRegistry.Range(func(name string, _ asynq.Handler) error {
		tasks = append(tasks, name)
		return nil
	})

	slices.Sort(tasks)
	for _, name := range tasks {
		fmt.Println(name)
	}

	return nil
}
