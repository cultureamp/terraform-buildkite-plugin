package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/outputs"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/validators"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/workingdir"
)

// Tests for getPluginName function

func TestGetPluginName(t *testing.T) {
	cfg := NewConfig().(*config)

	t.Run("valid URLs and names", func(t *testing.T) {
		cases := []struct {
			name  string
			input string
			want  string
		}{
			{
				name:  "github URL with version",
				input: "github.com/org/terraform-buildkite-plugin#v0.0.1",
				want:  "terraform-buildkite-plugin",
			},
			{
				name:  "simple plugin name",
				input: "terraform-buildkite-plugin",
				want:  "terraform-buildkite-plugin",
			},
			{
				name:  "https github URL",
				input: "https://github.com/org/terraform-buildkite-plugin#v0.0.1",
				want:  "terraform-buildkite-plugin",
			},
			{
				name:  "github URL without version",
				input: "github.com/org/terraform-buildkite-plugin",
				want:  "terraform-buildkite-plugin",
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				got := cfg.getPluginName(tt.input)
				assert.Equal(t, tt.want, got)
			})
		}
	})

	// Test cases that cover the error path (line 90: return s)
	t.Run("invalid URLs return original string", func(t *testing.T) {
		cases := []struct {
			name  string
			input string
			want  string
		}{
			{
				name:  "invalid URL scheme",
				input: "://invalid-url",
				want:  "://invalid-url",
			},
			{
				name:  "URL with invalid characters",
				input: "ht tp://invalid url",
				want:  "ht tp://invalid url",
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				got := cfg.getPluginName(tt.input)
				assert.Equal(t, tt.want, got)
			})
		}
	})
}

// Tests for unmarshalPlugins function

func TestUnmarshalPlugins(t *testing.T) {
	cfg := NewConfig().(*config)

	t.Run("valid plugin config", func(t *testing.T) {
		data := `[{"github.com/org/plugin#v0.0.1": {"mode": "plan"}}]`
		plugins, err := cfg.unmarshalPlugins(data)
		require.NoError(t, err)
		assert.Len(t, plugins, 1)
		assert.Contains(t, plugins[0], "github.com/org/plugin#v0.0.1")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		data := `invalid json`
		plugins, err := cfg.unmarshalPlugins(data)
		require.Error(t, err)
		assert.Nil(t, plugins)
		assert.Contains(t, err.Error(), "failed to parse plugin configuration")
	})

	t.Run("empty array", func(t *testing.T) {
		data := `[]`
		plugins, err := cfg.unmarshalPlugins(data)
		require.NoError(t, err)
		assert.Empty(t, plugins)
	})
}

// Tests for findRawPlugin function

func TestFindRawPlugin(t *testing.T) {
	cfg := NewConfig().(*config)

	t.Run("plugin found", func(t *testing.T) {
		data := []map[string]json.RawMessage{
			{"github.com/org/terraform-buildkite-plugin#v0.0.1": json.RawMessage(`{"mode": "plan"}`)},
		}
		plugin, err := cfg.findRawPlugin(data, "terraform-buildkite-plugin")
		require.NoError(t, err)
		assert.NotNil(t, plugin)
	})

	t.Run("plugin not found", func(t *testing.T) {
		data := []map[string]json.RawMessage{
			{"github.com/org/other-plugin#v0.0.1": json.RawMessage(`{"mode": "plan"}`)},
		}
		plugin, err := cfg.findRawPlugin(data, "terraform-buildkite-plugin")
		require.Error(t, err)
		assert.Nil(t, plugin)
		assert.Contains(t, err.Error(), "could not initialize plugin")
	})

	t.Run("empty plugin data", func(t *testing.T) {
		data := []map[string]json.RawMessage{}
		plugin, err := cfg.findRawPlugin(data, "terraform-buildkite-plugin")
		require.Error(t, err)
		assert.Nil(t, plugin)
	})
}

// Tests for parseRawPlugin function

func TestParseRawPlugin(t *testing.T) {
	cfg := NewConfig().(*config)

	t.Run("valid plugin configurations", func(t *testing.T) {
		t.Run("minimal config", func(t *testing.T) {
			data := json.RawMessage(`{"mode": "plan"}`)
			plugin, err := cfg.parseRawPlugin(&data)
			require.NoError(t, err)
			assert.NotNil(t, plugin)
			assert.Equal(t, Mode("plan"), plugin.Mode)
		})

		t.Run("complete config", func(t *testing.T) {
			data := json.RawMessage(`{
				"mode": "apply",
				"working": {
					"directory": "/path/to/terraform"
				},
				"outputs": [
					{
						"buildkite_annotation": {
							"template": "{{.output}}"
						}
					}
				]
			}`)
			plugin, err := cfg.parseRawPlugin(&data)
			require.NoError(t, err)
			assert.NotNil(t, plugin)
			assert.Equal(t, Mode("apply"), plugin.Mode)
			assert.NotNil(t, plugin.Working)
			assert.NotNil(t, plugin.Working.Directory)
			assert.Equal(t, "/path/to/terraform", *plugin.Working.Directory)
			assert.Len(t, plugin.Outputs.Outputs, 1)
			assert.NotNil(t, plugin.Outputs.Outputs[0].BuildkiteAnnotation)
			assert.Equal(t, "{{.output}}", plugin.Outputs.Outputs[0].BuildkiteAnnotation.Template)
		})
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("invalid JSON", func(t *testing.T) {
			bad := json.RawMessage(`{"bad":}`)
			plugin, err := cfg.parseRawPlugin(&bad)
			require.Error(t, err)
			assert.Nil(t, plugin)
			assert.Contains(t, err.Error(), "failed to parse plugin configuration")
		})

		// Test case that covers the environment variable parsing error path (lines 127-129)
		t.Run("environment variable parsing error", func(t *testing.T) {
			// Set an invalid environment variable that will cause env.Parse to fail
			// BUILDKITE_PARALLEL_JOB expects an integer, setting a non-numeric value should cause parsing to fail
			t.Setenv("BUILDKITE_PARALLEL_JOB", "not-a-number")

			data := json.RawMessage(`{"mode": "plan"}`)
			plugin, err := cfg.parseRawPlugin(&data)
			require.Error(t, err)
			assert.Nil(t, plugin)
			assert.Contains(t, err.Error(), "failed to parse environment variables")
		})
	})
}

// Tests for validatePlugin function

func TestValidatePlugin(t *testing.T) {
	cfg := NewConfig().(*config)

	t.Run("valid configurations", func(t *testing.T) {
		t.Run("minimal config", func(t *testing.T) {
			plugin := &Plugin{
				Mode: "plan",
			}
			err := cfg.validatePlugin(plugin)
			require.NoError(t, err)
		})

		t.Run("complete config", func(t *testing.T) {
			workingDir := t.TempDir()
			plugin := &Plugin{
				Mode: "apply",
				Working: &workingdir.Working{
					Directory: &workingDir,
				},
				Validations: validators.Validations{
					Validations: []validators.Validation{
						{
							Opa: &validators.OpaValidation{
								Bundle: "policy.tar.gz",
								Query:  "terraform/allow",
							},
						},
					},
				},
				Outputs: outputs.Outputs{
					Outputs: []outputs.Output{{
						BuildkiteAnnotation: &outputs.BuildkiteAnnotation{
							Template: "{{.output}}",
							Context:  "terraform-output",
						},
					}},
				},
			}
			err := cfg.validatePlugin(plugin)
			require.NoError(t, err)
		})

		t.Run("working directories config", func(t *testing.T) {
			parentDir := t.TempDir()
			plugin := &Plugin{
				Mode: "plan",
				Working: &workingdir.Working{
					Directories: &workingdir.Directories{
						ParentDirectory: parentDir,
						NameRegex:       ".*",
					},
				},
			}
			err := cfg.validatePlugin(plugin)
			require.NoError(t, err)
		})
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Run("missing required mode", func(t *testing.T) {
			plugin := &Plugin{} // missing required mode field
			err := cfg.validatePlugin(plugin)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to validate config")
		})

		t.Run("invalid mode", func(t *testing.T) {
			plugin := &Plugin{
				Mode: "invalid",
			}
			err := cfg.validatePlugin(plugin)
			require.Error(t, err)
		})

		t.Run("both working_directory and working_directories set", func(t *testing.T) {
			workingDir := t.TempDir()
			parentDir := t.TempDir()
			plugin := &Plugin{
				Mode: Plan,
				Working: &workingdir.Working{
					Directory: &workingDir,
					Directories: &workingdir.Directories{
						ParentDirectory: parentDir,
					},
				},
			}
			err := cfg.validatePlugin(plugin)
			require.Error(t, err)
		})
	})
}
