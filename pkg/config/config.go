// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	coreconfig "github.com/gardener/inventory/pkg/core/config"
)

// ConfigFormatVersion represents the supported config format version for the
// extension.
const ConfigFormatVersion = "v1alpha1"

// Config represents the extension configuration.
type Config struct {
	// Version is the version of the config file.
	Version string `yaml:"version"`

	// Debug configures debug mode, if set to true.
	Debug bool `yaml:"debug"`

	// Logging provides the logging config settings.
	Logging coreconfig.LoggingConfig `yaml:"logging"`

	// Redis provides the Redis configuration.
	Redis coreconfig.RedisConfig `yaml:"redis"`

	// Database provides the database configuration.
	Database coreconfig.DatabaseConfig `yaml:"database"`

	// Worker provides the worker configuration.
	Worker coreconfig.WorkerConfig `yaml:"worker"`
}

// Parse parses the configs from the given paths in-order. Configuration
// settings provided later in the sequence of paths will override settings from
// previous config paths.
func Parse(paths []string) (*Config, error) {
	var conf Config

	for _, path := range paths {
		if err := coreconfig.ParseFileInto(path, &conf); err != nil {
			return nil, err
		}

		if conf.Version == "" {
			return nil, fmt.Errorf("%w: %s", coreconfig.ErrNoConfigVersion, path)
		}

		if conf.Version != ConfigFormatVersion {
			return nil, fmt.Errorf("%w: %s (%s)", coreconfig.ErrUnsupportedVersion, conf.Version, path)
		}

	}

	return &conf, nil
}
