// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

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
	odgapi "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/api/client"
	odgclient "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/client"
)

// NewWorkerCommand returns a new [cli.Command] for worker-related operations.
func NewWorkerCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "worker",
		Usage:   "worker operations",
		Aliases: []string{"w"},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "config",
				Usage:    "path to extension config file",
				Required: true,
				Aliases:  []string{"file"},
				EnvVars:  []string{"INVENTORY_EXTENSION_CONFIG"},
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "start worker process",
				Aliases: []string{"s"},
				Action:  execWorkerStartCommand,
			},
			{
				Name:    "ping",
				Usage:   "ping a worker",
				Aliases: []string{"p"},
				Action:  execWorkerPingCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "worker",
						Usage:    "worker name to ping",
						Required: true,
						Aliases:  []string{"name"},
					},
				},
			},
		},
	}

	return cmd
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

// newOdgClient creates a new [odgapiclient.Client] instance based on the
// provided [config.Config] settings.
func newOdgClient(conf *config.Config) (*odgapi.Client, error) {
	opts := []odgapi.Option{
		odgapi.WithUserAgent(conf.ODG.UserAgent),
	}

	if conf.ODG.Endpoint == "" {
		return nil, errors.New("odg: no api endpoint specified")
	}

	if conf.ODG.Auth.Method == "" {
		return nil, errors.New("odg: no auth method specified")
	}

	switch conf.ODG.Auth.Method {
	case config.ODGAuthMethodGithub:
		if conf.ODG.Auth.Github.URL == "" {
			return nil, errors.New("odg: no github api url specified")
		}
		if conf.ODG.Auth.Github.Token == "" {
			return nil, errors.New("odg: no github access token specified")
		}
		opts = append(
			opts,
			odgapi.WithGithubAuthentication(conf.ODG.Auth.Github.URL, conf.ODG.Auth.Github.Token),
		)
	case config.ODGAuthMethodNone:
		// No authentication, nothing to do here.
	default:
		return nil, fmt.Errorf("odg: unknown auth method %s", conf.ODG.Auth.Method)
	}

	return odgapi.New(conf.ODG.Endpoint, opts...)
}

// execWorkerStartCommand starts the worker
func execWorkerStartCommand(ctx *cli.Context) error {
	// Parse config files for the extension
	configPaths := ctx.StringSlice("config")
	conf, err := config.Parse(configPaths...)
	if err != nil {
		return err
	}

	// Configure the default [slog.Logger]
	logger, err := slogutils.NewFromConfig(os.Stdout, conf.Logging)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	slog.Info("configured default logger")

	// Configure database client and set it up for task handlers
	slog.Info("configuring database client")
	db, err := newDB(conf)
	if err != nil {
		return err
	}
	dbclient.SetDB(db)
	defer db.Close()

	// Configure the Open Delivery Gear API client
	slog.Info(
		"configuring open delivery gear api client",
		"endpoint", conf.ODG.Endpoint,
		"auth", conf.ODG.Auth.Method,
	)
	odgClient, err := newOdgClient(conf)
	if err != nil {
		return err
	}

	if conf.ODG.Auth.Method != config.ODGAuthMethodNone {
		if err := odgClient.Authenticate(ctx.Context); err != nil {
			return err
		}
		defer func() {
			_ = odgClient.Logout(ctx.Context)
		}()
	}
	odgclient.SetClient(odgClient)

	// Create a worker, register handlers and start it up
	worker := newWorker(conf)
	worker.HandlersFromRegistry(registry.TaskRegistry)
	_ = registry.TaskRegistry.Range(func(name string, _ asynq.Handler) error {
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

// execWorkerPingCommand pings a worker
func execWorkerPingCommand(ctx *cli.Context) error {
	// Parse config files for the extension
	configPaths := ctx.StringSlice("config")
	conf, err := config.Parse(configPaths...)
	if err != nil {
		return err
	}

	workerName := ctx.String("worker")
	redisClientOpt := asynqutils.NewRedisClientOptFromConfig(conf.Redis)
	inspector := asynq.NewInspector(redisClientOpt)
	defer inspector.Close()

	servers, err := inspector.Servers()
	if err != nil {
		return err
	}

	ok := false
	for _, server := range servers {
		if server.Host == workerName {
			ok = true
			fmt.Printf("%s/%d: OK\n", server.Host, server.PID)
		}
	}

	if !ok {
		return cli.Exit("", 1)
	}

	return nil
}
