// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	_ "github.tools.sap/kubernetes/inventory-extension-odg/pkg/odg/tasks"
	"github.tools.sap/kubernetes/inventory-extension-odg/pkg/version"
)

func main() {
	app := &cli.App{
		Name:                 "inventory-extension-odg",
		Version:              version.Version,
		EnableBashCompletion: true,
		Usage:                "inventory extension for open delivery gear",
		Commands: []*cli.Command{
			NewWorkerCommand(),
			NewTasksCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
