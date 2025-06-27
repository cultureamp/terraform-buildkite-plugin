package initiator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/outputs"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/validators"
	c "github.com/xphir/terraform-buildkite-plugin/internal/config"
)

type ParsedPayload struct {
	Plugin             *c.Plugin
	Outputers          []outputs.Outputer
	Validators         []validators.Validator
	WorkingDirectories []string
}

type PluginInitiator interface {
	ParsePlugin(ctx context.Context, pluginName string) (*ParsedPayload, error)
}

type initiatorConfig struct {
	configInterface c.Config // The raw plugin configuration
}

type Option func(*initiatorConfig)

func WithConfigInterface(c c.Config) Option {
	return func(r *initiatorConfig) {
		if c != nil {
			r.configInterface = c
		}
	}
}

// NewInitiator creates a new instance of the plugin with the provided configuration options.
func NewInitiator(opts ...Option) PluginInitiator {
	defaults := &initiatorConfig{
		configInterface: c.NewConfig(),
	}
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}

func (i *initiatorConfig) ParsePlugin(
	ctx context.Context,
	pluginName string,
) (*ParsedPayload, error) {
	log.Info().Msg("loading and parsing plugin configuration")
	// Initialize plugin configuration
	plugin, err := i.configInterface.LoadPlugin(ctx, pluginName)
	if err != nil {
		log.Error().Str("plugin", pluginName).Msg("failed to initialize plugin")
		return nil, err
	}
	outputers, err := plugin.Outputs.ToOutputers()
	if err != nil {
		log.Error().Err(err).Msg("failed to convert outputs to outputers")
		return nil, fmt.Errorf("failed to convert outputs: %w", err)
	}
	validators, err := plugin.Validations.ToValidators()
	if err != nil {
		log.Error().Err(err).Msg("failed to convert validations to validators")
		return nil, fmt.Errorf("failed to convert validations: %w", err)
	}
	dirs, err := plugin.Working.Parse()
	if err != nil {
		log.Error().Err(err).Msg("failed to parse working directories")
		return nil, fmt.Errorf("failed to parse working directories: %w", err)
	}
	log.Info().Msg("plugin configuration loaded and parsed successfully")
	return &ParsedPayload{plugin, outputers, validators, dirs}, nil
}
