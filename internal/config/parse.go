package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/workingdir"
)

// getPluginName extracts the repository name from a plugin reference string.
func (c *config) getPluginName(s string) string {
	ref := s
	if strings.HasPrefix(ref, "github.com/") && !strings.Contains(ref, "://") {
		ref = "https://" + ref
	}
	u, err := url.Parse(ref)
	if err != nil {
		log.Debug().Str("input", s).Msg("failed to parse plugin reference as URL")
		return s
	}
	_, repo := path.Split(u.Path)
	return repo
}

// unmarshalPlugins parses the Buildkite plugins JSON array.
func (c *config) unmarshalPlugins(data string) ([]map[string]json.RawMessage, error) {
	log.Debug().Msg("parsing plugin configuration JSON")
	var pluginConfigs []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &pluginConfigs); err != nil {
		log.Error().Msg("failed to unmarshal plugin configuration JSON")
		return nil, errors.New("failed to parse plugin configuration")
	}
	return pluginConfigs, nil
}

// findRawPlugin returns the first plugin config whose key matches pluginName prefix.
func (c *config) findRawPlugin(data []map[string]json.RawMessage, pluginName string) (*json.RawMessage, error) {
	for _, p := range data {
		for key, pluginConfig := range p {
			if strings.HasPrefix(c.getPluginName(key), pluginName) {
				log.Debug().Str("matched_key", key).Msg("found matching plugin configuration")
				return &pluginConfig, nil
			}
		}
	}
	log.Error().Str("plugin", pluginName).Msg("could not find matching plugin configuration")
	return nil, errors.New("could not initialize plugin")
}

// parseRawPlugin loads a Plugin from environment variables and overlays JSON config.
func (c *config) parseRawPlugin(data *json.RawMessage) (*Plugin, error) {
	log.Debug().Msg("parsing environment variables into Plugin struct")

	var plugin Plugin

	// Pre-initialize nested pointer structs if environment variables are present
	// The env library needs these structs to exist before it can populate them
	if os.Getenv("BUILDKITE_PARALLEL_JOB") != "" || os.Getenv("BUILDKITE_PARALLEL_JOB_COUNT") != "" {
		plugin.Working = &workingdir.Working{
			Parallelism: &workingdir.Parallelism{},
		}
	}

	if err := env.Parse(&plugin); err != nil {
		log.Error().Msg("failed to parse environment variables")
		return nil, errors.New("failed to parse environment variables")
	}

	log.Debug().Msg("applying JSON overrides to Plugin struct")
	if err := json.Unmarshal(*data, &plugin); err != nil {
		log.Error().Msg("failed to unmarshal plugin JSON")
		return nil, errors.New("failed to parse plugin configuration")
	}

	log.Debug().Msg("plugin configuration parsed successfully")
	return &plugin, nil
}

// validatePlugin checks struct tags and field constraints.
func (c *config) validatePlugin(plugin *Plugin) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(plugin); err != nil {
		log.Error().Msg("plugin validation failed")
		return fmt.Errorf("failed to validate config: %w", err)
	}
	return nil
}
