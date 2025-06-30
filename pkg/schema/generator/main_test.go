package generator_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/generator"
	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/handler"
	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/schema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// mockSchemaConfig implements schema.SchemaConfig for testing.
type mockSchemaConfig struct {
	shouldError bool
}

func (m *mockSchemaConfig) GeneratePluginSchema() (*schema.PluginSchema, error) {
	if m.shouldError {
		return nil, errors.New("mock schema error")
	}
	return &schema.PluginSchema{
		Name:          "test",
		Description:   "desc",
		Author:        "author",
		Requirements:  []string{"req"},
		Configuration: map[string]any{"foo": "bar"},
	}, nil
}

func TestGenerator_TableDriven(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.yaml")

	tests := []struct {
		name           string
		handler        *handler.MockHandler
		outputFile     string
		expectErr      bool
		expectContains string
	}{
		{
			name: "Success",
			handler: &handler.MockHandler{
				HandleFunc: func(_ schema.Config, _ *handler.HandleOptions) func(cmd *cobra.Command, args []string) error {
					return func(_ *cobra.Command, _ []string) error {
						return nil // Don't write to a file
					}
				},
			},
			outputFile:     outputFile,
			expectErr:      false,
			expectContains: "",
		},
		{
			name: "Handler error",
			handler: &handler.MockHandler{
				HandleReturnErr: errors.New("handler error"),
			},
			outputFile:     "ignored.yaml",
			expectErr:      true,
			expectContains: "handler error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &handler.HandleOptions{OutputFile: tt.outputFile}
			cmd := &cobra.Command{}
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			command := func(runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
				cmd.RunE = runE
				cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", tt.outputFile, "Output file")
				return cmd
			}

			gen := generator.New(generator.WithCommand(command), generator.WithHandler(tt.handler))
			err := gen.GenerateSchema(t.Context(), &mockSchemaConfig{})
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerator_DefaultCommandIsSet(t *testing.T) {
	handler := &handler.MockHandler{
		HandleFunc: func(_ schema.Config, _ *handler.HandleOptions) func(cmd *cobra.Command, args []string) error {
			return func(_ *cobra.Command, _ []string) error {
				return nil
			}
		},
	}
	gen := generator.New(generator.WithHandler(handler))
	require.NotNil(t, gen, "generator should not be nil")

	err := gen.GenerateSchema(t.Context(), &mockSchemaConfig{})
	require.NoError(t, err, "default command should be set and callable")
}
