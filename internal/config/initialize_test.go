package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/outputs"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/validators"
	"github.com/xphir/terraform-buildkite-plugin/internal/adapters/workingdir"
	"github.com/xphir/terraform-buildkite-plugin/internal/config"
)

// Tests for NewConfig function.
func TestNewConfigOptions(t *testing.T) {
	t.Run("supports option functions", func(t *testing.T) {
		cfg := config.NewConfig()
		assert.NotNil(t, cfg)
	})
}

// Tests for WithPluginsEnv function.
func TestWithPluginsEnv(t *testing.T) {
	t.Run("applies environment override", func(t *testing.T) {
		pluginConfig := `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {"mode": "plan"}}]`
		t.Setenv("CUSTOM_PLUGINS_ENV", pluginConfig)

		cfg := config.NewConfig(config.WithPluginsEnv("CUSTOM_PLUGINS_ENV"))
		plugin, err := cfg.LoadPlugin(
			t.Context(),
			"terraform-buildkite-plugin",
		)

		require.NoError(t, err)
		assert.NotNil(t, plugin)
		expected := &config.Plugin{Mode: config.Plan}
		assert.Equal(t, expected, plugin)
	})
}

// Tests for InitializePlugin function.
func TestInitializePlugin(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		t.Run("complete configuration with all fields", func(t *testing.T) {
			// Create a temporary directory for the test
			workingDir := t.TempDir()

			pluginConfig := `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {
				"mode": "apply",
				"working": {
					"directory": "` + workingDir + `"
				},
				"validations": [
					{ "opa": { "bundle": "policy.tar.gz", "query": "terraform/allow" } }
				],
				"outputs": [
					{
						"buildkite_annotation": {
							"template": "{{.output}}",
							"context": "terraform-output",
							"vars": [{"key": "value"}],
							"computed_vars": [
								{ "name": "namespace", "from": "working_dir", "regex": "^[^.]+\\.(.+)\\.[^.]+$" }
							]
						}
					}
				]
			}}]`
			t.Setenv("BUILDKITE_PLUGINS", pluginConfig)

			cfg := config.NewConfig()
			plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

			require.NoError(t, err)
			assert.NotNil(t, plugin)

			expected := &config.Plugin{
				Mode: config.Apply,
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
							Vars:     []map[string]string{{"key": "value"}},
							ComputedVars: []outputs.ComputedVar{
								{
									Name:  "namespace",
									From:  "working_dir",
									Regex: "^[^.]+\\.(.+)\\.[^.]+$",
								},
							},
						}},
					},
				},
			}
			assert.Equal(t, expected, plugin)
		})

		t.Run("plugin with complex output configuration", func(t *testing.T) {
			pluginConfig := `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {
				"mode": "plan",
				"outputs": [
					{
						"buildkite_annotation": {
							"template": "Plan result: {{.plan_result}}\nChanges: {{.changes}}",
							"context": "terraform-plan-output",
							"vars": [
								{"environment": "production"},
								{"region": "us-west-2"}
							],
							"computed_vars": [
								{"name": "workspace", "from": "working_directory", "regex": "/([^/]+)/?$"},
								{"name": "environment", "from": "workspace_name", "regex": "^env-(.+)$"}
							]
						}
					}
				]
			}}]`
			t.Setenv("BUILDKITE_PLUGINS", pluginConfig)

			cfg := config.NewConfig()
			plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

			require.NoError(t, err)
			require.NotNil(t, plugin)

			expected := &config.Plugin{
				Mode: config.Plan,
				Outputs: outputs.Outputs{
					Outputs: []outputs.Output{{
						BuildkiteAnnotation: &outputs.BuildkiteAnnotation{
							Template: "Plan result: {{.plan_result}}\nChanges: {{.changes}}",
							Context:  "terraform-plan-output",
							Vars:     []map[string]string{{"environment": "production"}, {"region": "us-west-2"}},
							ComputedVars: []outputs.ComputedVar{
								{
									Name:  "workspace",
									From:  "working_directory",
									Regex: "/([^/]+)/?$",
								},
								{
									Name:  "environment",
									From:  "workspace_name",
									Regex: "^env-(.+)$",
								},
							},
						}},
					},
				},
			}

			assert.Equal(t, expected, plugin)
		})
	})

	t.Run("environment variable overrides", func(t *testing.T) {
		t.Run("environment variables override JSON defaults", func(t *testing.T) {
			pluginConfig := `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {"mode": "plan"}}]`
			t.Setenv("BUILDKITE_PLUGINS", pluginConfig)
			t.Setenv("BUILDKITE_PARALLEL_JOB", "2")
			t.Setenv("BUILDKITE_PARALLEL_JOB_COUNT", "5")

			cfg := config.NewConfig()
			plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

			require.NoError(t, err)
			assert.NotNil(t, plugin)

			parallelJob := 2
			parallelJobCount := 5
			expected := &config.Plugin{
				Mode: config.Plan,
				Working: &workingdir.Working{
					Parallelism: &workingdir.Parallelism{
						ParallelJob:      &parallelJob,
						ParallelJobCount: &parallelJobCount,
					},
				},
			}
			assert.Equal(t, expected, plugin)
		})

		t.Run("JSON overrides environment variables", func(t *testing.T) {
			pluginConfig := `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {
				"mode": "plan"
			}}]`
			t.Setenv("BUILDKITE_PLUGINS", pluginConfig)
			t.Setenv("LOG_LEVEL", "debug")

			cfg := config.NewConfig()
			plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

			require.NoError(t, err)
			assert.NotNil(t, plugin)

			expected := &config.Plugin{
				Mode: "plan",
			}
			assert.Equal(t, expected, plugin)
		})
	})

	t.Run("error cases", func(t *testing.T) {
		cases := []struct {
			name          string
			envValue      string
			expectedError string
		}{
			{"missing environment variable", "", "failed to parse plugin configuration"},
			{"invalid JSON", "not-json", "failed to parse plugin configuration"},
			{"malformed JSON", `[{"key": }]`, "failed to parse plugin configuration"},
			{
				"plugin not found",
				`[{"github.com/org/other-plugin#v0.0.1": {"mode": "plan"}}]`,
				"could not initialize plugin",
			},
			{"empty plugin array", "[]", "could not initialize plugin"},
			{
				"invalid plugin structure (parseRawPlugin fails)",
				`[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {"mode": {"invalid": "structure"}}}]`,
				"failed to parse plugin configuration",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.envValue != "" {
					t.Setenv("BUILDKITE_PLUGINS", tc.envValue)
				}

				cfg := config.NewConfig()
				plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

				require.Error(t, err)
				assert.Nil(t, plugin)
				assert.Contains(t, err.Error(), tc.expectedError)
			})
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		cases := []struct {
			name          string
			pluginConfig  string
			expectedError string
		}{
			{
				"missing required mode",
				`[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {}}]`,
				"failed to validate config",
			},
			{
				"invalid mode",
				`[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {"mode": "invalid"}}]`,
				"failed to validate config",
			},
			{"both working_directory and working_directories", `[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {
				"mode": "plan",
				"working": {
					"directory": "/path",
					"directories": {"parent_directory": "/parent"}
				}
			}}]`, "failed to validate config"},
			{
				"working_directories with both parent_directory and artifact",
				`[{"github.com/org/terraform-buildkite-plugin#v0.0.1": {
				"mode": "plan",
				"working": {
					"directories": {
						"parent_directory": "/parent",
						"artifact": "configs.tar.gz"
					}
				}
			}}]`,
				"failed to validate config",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Setenv("BUILDKITE_PLUGINS", tc.pluginConfig)

				cfg := config.NewConfig()
				plugin, err := cfg.LoadPlugin(t.Context(), "terraform-buildkite-plugin")

				require.Error(t, err)
				assert.Nil(t, plugin)
				assert.Contains(t, err.Error(), tc.expectedError)
			})
		}
	})
}
