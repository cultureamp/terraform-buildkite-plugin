package schema

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

// PluginSchema represents the structure for the plugin metadata and configuration.
type PluginSchema struct {
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Author        string         `yaml:"author"`
	Requirements  []string       `yaml:"requirements"`
	Configuration map[string]any `yaml:"configuration"`
}

type Config interface {
	GeneratePluginSchema() (*PluginSchema, error)
}

type PluginProperties struct {
	Name         string
	Description  string
	Author       string
	Requirements []string
}

type config struct {
	Properties *PluginProperties
	Schema     any
	Caller     string
}

type ConfigOption func(*config)

func WithProperties(p *PluginProperties) ConfigOption {
	return func(r *config) {
		if p != nil {
			r.Properties = p
		}
	}
}

func WithSchema(s any) ConfigOption {
	return func(r *config) {
		if s != nil {
			r.Schema = s
		}
	}
}

func WithCaller(c string) ConfigOption {
	return func(r *config) {
		r.Caller = c
	}
}

func New(opts ...ConfigOption) Config {
	config := &config{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type JSONSchema map[string]any

func GenerateJSONSchema(input any) (JSONSchema, error) {
	if err := validateInputSchema(input); err != nil {
		return nil, err
	}
	ref := &jsonschema.Reflector{
		DoNotReference: true,
		Anonymous:      true,
	}

	schema := ref.Reflect(input)
	schema.Version = "" // Clear the version if not needed

	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	var result map[string]any
	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON schema: %w", err)
	}

	return result, nil
}

func (g *config) GeneratePluginSchema() (*PluginSchema, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	configuration, err := GenerateJSONSchema(g.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate configuration schema: %w", err)
	}

	return &PluginSchema{
		Name:          g.Properties.Name,
		Description:   g.Properties.Description,
		Author:        g.Properties.Author,
		Requirements:  g.Properties.Requirements,
		Configuration: configuration,
	}, nil
}
