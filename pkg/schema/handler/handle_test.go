package handler_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/caller"
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

func TestHandle_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.yaml")

	h := handler.New(
		handler.WithCaller(&caller.MockCaller{
			CallPathResult: "./mock",
			CallPathErr:    nil,
		}),
	)
	opts := &handler.HandleOptions{OutputFile: outputFile}
	mockConfig := &mockSchemaConfig{}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	runE := h.Handle(mockConfig, opts)
	err := runE(cmd, []string{})
	require.NoError(t, err)
	out := buf.String()
	require.Contains(t, out, "Generating plugin schema to "+outputFile)
	require.Contains(t, out, "✅ Plugin schema successfully generated and saved to "+outputFile)

	// Ensure the file was actually written
	contents, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	require.Contains(t, string(contents), "test")
}

func TestHandle_ValidateOptionsError(t *testing.T) {
	h := handler.New(
		handler.WithCaller(&caller.MockCaller{
			CallPathResult: "./mock",
			CallPathErr:    nil,
		}),
	)
	opts := &handler.HandleOptions{OutputFile: ""}
	mockConfig := &mockSchemaConfig{}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	runE := h.Handle(mockConfig, opts)
	err := runE(cmd, []string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to validate options")
}

func TestHandle_GeneratePluginSchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.yaml")

	h := handler.New(
		handler.WithCaller(&caller.MockCaller{
			CallPathResult: "./mock",
			CallPathErr:    nil,
		}),
	)
	opts := &handler.HandleOptions{OutputFile: outputFile}
	mockConfig := &mockSchemaConfig{shouldError: true}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	runE := h.Handle(mockConfig, opts)
	err := runE(cmd, []string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "error generating plugin schema")
}
