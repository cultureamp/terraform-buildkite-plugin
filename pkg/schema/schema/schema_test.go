package schema_test

import (
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/schema"
	"github.com/stretchr/testify/require"
)

type ValidConfig struct {
	Field string `json:"field"`
}

type InvalidConfig func() // cannot reflect or marshal

func TestGenerateJSONSchema(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		mockUnmarshal bool
		expectErr     bool
		errMsg        string
	}{
		{
			name:      "Valid struct",
			input:     ValidConfig{Field: "value"},
			expectErr: false,
		},
		{
			name:      "Invalid input: int",
			input:     123,
			expectErr: true,
			errMsg:    "expected struct, got int",
		},
		{
			name:      "Invalid input: string",
			input:     "string",
			expectErr: true,
			errMsg:    "expected struct, got string",
		},
		{
			name:      "Invalid input: slice",
			input:     []string{"a", "b"},
			expectErr: true,
			errMsg:    "expected struct, got []string",
		},
		{
			name:      "Nil input",
			input:     nil,
			expectErr: true,
			errMsg:    "input cannot be nil",
		},
		{
			name:      "Pointer to struct",
			input:     &ValidConfig{Field: "value"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := schema.GenerateJSONSchema(tt.input)
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schema)
			}
		})
	}
}

// --- Tests for GeneratorConfig.GenerateSchema ---

func TestGeneratorConfig_GenerateSchema(t *testing.T) {
	tests := []struct {
		name         string
		config       schema.Config
		expectErr    bool
		errMsg       string
		expectedName string
	}{
		{
			name: "Valid configuration",
			config: schema.New(
				schema.WithSchema(ValidConfig{Field: "value"}),
				schema.WithProperties(&schema.PluginProperties{
					Name:         "TestPlugin",
					Description:  "Test description",
					Author:       "TestAuthor",
					Requirements: []string{"req1", "req2"},
				}),
			),
			expectErr:    false,
			expectedName: "TestPlugin",
		},
		{
			name: "Invalid schema",
			config: schema.New(
				schema.WithSchema(123),
				schema.WithProperties(&schema.PluginProperties{
					Name:         "TestPlugin",
					Description:  "Test description",
					Author:       "TestAuthor",
					Requirements: []string{"req1", "req2"},
				}),
			),
			expectErr: true,
			errMsg:    "expected struct, got int",
		},
		{
			name: "Nil properties",
			config: schema.New(
				schema.WithSchema(ValidConfig{Field: "value"}),
			),
			expectErr: true,
			errMsg:    "properties cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginSchema, err := tt.config.GeneratePluginSchema()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pluginSchema)
				require.Equal(t, tt.expectedName, pluginSchema.Name)
			}
		})
	}
}
