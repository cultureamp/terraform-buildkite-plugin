// Package config provides configuration management for the Terraform Buildkite plugin.
//
// It supports loading, parsing, and validating plugin configuration from environment variables and JSON sources.
// JSON config is loaded from the BUILDKITE_PLUGINS environment variable and overlays environment values.
//
// Validation includes required fields, value constraints, and mutual exclusion rules.
package config

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/xphir/terraform-buildkite-plugin/internal/common"
)

// Config provides methods for loading and validating Buildkite plugin configurations.
type Config interface {
	// LoadPlugin loads, parses, and validates a named Buildkite plugin configuration.
	LoadPlugin(ctx context.Context, pluginName string) (*Plugin, error)
}

// config implements Config.
type config struct {
	pluginsEnv string // Environment variable name for plugin JSON configuration
}

// Option configures a Config instance during creation.
type Option func(*config)

// WithPluginsEnv sets a custom environment variable for plugin JSON.
func WithPluginsEnv(n string) Option {
	return func(ic *config) {
		ic.pluginsEnv = n
	}
}

// NewConfig creates a new Config instance.
func NewConfig(opts ...Option) Config {
	c := &config{
		pluginsEnv: "BUILDKITE_PLUGINS",
	}
	for _, o := range opts {
		o(c)
	}
	log.Debug().Msg("config instance created")
	return c
}

// LoadPlugin loads, parses, and validates a Buildkite plugin configuration.
func (c *config) LoadPlugin(_ context.Context, pluginName string) (*Plugin, error) {
	log.Debug().Str("plugin", pluginName).Msg("initializing plugin configuration")

	raw := common.FetchEnv(c.pluginsEnv, "")
	entries, err := c.unmarshalPlugins(raw)
	if err != nil {
		log.Error().Str("plugin", pluginName).Msg("failed to unmarshal plugins JSON")
		return nil, err
	}

	var rawEntry *json.RawMessage
	rawEntry, err = c.findRawPlugin(entries, pluginName)
	if err != nil {
		log.Error().Str("plugin", pluginName).Msg("plugin not found in configuration")
		return nil, err
	}

	var plugin *Plugin
	plugin, err = c.parseRawPlugin(rawEntry)
	if err != nil {
		log.Error().Str("plugin", pluginName).Msg("failed to parse plugin config")
		return nil, err
	}

	if err = c.validatePlugin(plugin); err != nil {
		log.Error().Str("plugin", pluginName).Msg("plugin config validation failed")
		return nil, err
	}

	log.Info().Str("plugin", pluginName).Msg("plugin initialized successfully")
	log.Debug().Interface("plugin", plugin).Msg("plugin configuration details")
	return plugin, nil
}
