// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	coreconfig "github.com/gardener/inventory/pkg/core/config"
)

// ODGAuthMethod represents an authentication method to use when authenticating
// against the remote Delivery Service API.
type ODGAuthMethod string

const (
	// ODGAuthMethodGithub represents authentication method, which uses
	// Github for querying users' information.
	ODGAuthMethodGithub = "github"

	// ODGAuthMethodNone is the name of the method, in which the API client
	// will use no authentication against the remote API service.
	ODGAuthMethodNone = "none"
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

	// ODG provides the Open Delivery Gear configuration
	ODG ODGConfig `yaml:"odg"`
}

// ODGConfig represents the Open Delivery Gear configuration
type ODGConfig struct {
	// Endpoint specifies the base API endpoint of the remote API
	Endpoint string `yaml:"endpoint"`

	// UserAgent specifies the User-Agent header to configure for the API
	// client.
	UserAgent string `yaml:"user_agent"`

	Auth ODGAuthConfig `yaml:"auth"`
}

// ODGAuthConfig represents the Open Delivery Gear authentication configuration.
type ODGAuthConfig struct {
	// Method specifies the authentication method to use when authenticating
	// against the remote Open Delivery Gear API.
	Method ODGAuthMethod `yaml:"method"`

	// Github specifies the settings for `github' authentication method when
	// authenticating against the remote API.
	Github ODGAuthGithubConfig `yaml:"github"`
}

// ODGAuthGithubConfig provides the configuration for `github' authentication
// method.
type ODGAuthGithubConfig struct {
	// URL specifies the base Github API URL which the Delivery Service will
	// use to query user's information with the provided access token.
	URL string `yaml:"url"`

	// Token specifies the Github access token which will be used to query
	// the information about the user associated with the token.
	Token string `yaml:"token"`
}

// Parse parses the configs from the given paths in-order. Configuration
// settings provided later in the sequence of paths will override settings from
// previous config paths.
func Parse(paths ...string) (*Config, error) {
	var conf Config

	for _, path := range paths {
		// Ignore empty paths
		if path == "" {
			continue
		}

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
