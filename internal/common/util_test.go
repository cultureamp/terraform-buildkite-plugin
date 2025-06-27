package common_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xphir/terraform-buildkite-plugin/internal/common"
)

func TestFetchEnv(t *testing.T) {
	t.Run("returns environment variable when set", func(t *testing.T) {
		key := "TEST_ENV_VAR"
		expectedValue := "test_value"
		fallback := "fallback_value"

		// Set environment variable
		t.Setenv(key, expectedValue)

		result := common.FetchEnv(key, fallback)
		assert.Equal(t, expectedValue, result)
	})

	t.Run("returns fallback when environment variable not set", func(t *testing.T) {
		key := "NON_EXISTENT_ENV_VAR"
		fallback := "fallback_value"

		// Ensure the environment variable is not set
		os.Unsetenv(key)

		result := common.FetchEnv(key, fallback)
		assert.Equal(t, fallback, result)
	})

	t.Run("returns empty string when env var is empty and fallback is empty", func(t *testing.T) {
		key := "EMPTY_ENV_VAR"
		fallback := ""

		// Set environment variable to empty string
		t.Setenv(key, "")

		result := common.FetchEnv(key, fallback)
		assert.Empty(t, result)
	})

	t.Run("returns environment variable over fallback when both exist", func(t *testing.T) {
		key := "PRIORITY_TEST_VAR"
		envValue := "env_value"
		fallback := "fallback_value"

		// Set environment variable
		t.Setenv(key, envValue)

		result := common.FetchEnv(key, fallback)
		assert.Equal(t, envValue, result)
		assert.NotEqual(t, fallback, result)
	})
}

func TestWritePrettyJSON(t *testing.T) {
	t.Run("prints valid JSON", func(t *testing.T) {
		data := map[string]any{"foo": "bar", "num": 42}
		buf := &bytes.Buffer{}

		err := common.WritePrettyJSON(data, buf)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "foo")
		assert.Contains(t, buf.String(), "bar")
		assert.Contains(t, buf.String(), "num")
	})

	t.Run("errors on invalid JSON", func(t *testing.T) {
		data := map[string]any{"bad": make(chan int)}
		buf := &bytes.Buffer{}
		err := common.WritePrettyJSON(data, buf)
		require.Error(t, err)
	})
}

func TestSetLogLevel(t *testing.T) {
	t.Run("sets valid log levels", func(t *testing.T) {
		levels := map[string]zerolog.Level{
			"trace": zerolog.TraceLevel,
			"debug": zerolog.DebugLevel,
			"info":  zerolog.InfoLevel,
			"warn":  zerolog.WarnLevel,
			"error": zerolog.ErrorLevel,
			"fatal": zerolog.FatalLevel,
			"panic": zerolog.PanicLevel,
		}
		for levelStr, want := range levels {
			common.SetLogLevel(levelStr)
			got := zerolog.GlobalLevel()
			assert.Equal(t, want, got, "log level %s should set zerolog.GlobalLevel() to %d", levelStr, want)
		}
	})

	t.Run("defaults to info on invalid level", func(t *testing.T) {
		common.SetLogLevel("notalevel")
		got := int(zerolog.GlobalLevel())
		assert.Equal(t, int(zerolog.InfoLevel), got)
	})
}

func TestIsTestMode(t *testing.T) {
	plugin := "terraform-buildkite-plugin"
	varName := "BUILDKITE_PLUGIN_TERRAFORM_BUILDKITE_PLUGIN_TEST_MODE"

	t.Run("returns true for 'true' value", func(t *testing.T) {
		t.Setenv(varName, "true")
		assert.True(t, common.IsTestMode(plugin))
	})

	t.Run("returns true for '1' value", func(t *testing.T) {
		t.Setenv(varName, "1")
		assert.True(t, common.IsTestMode(plugin))
	})

	t.Run("returns true for 'yes' value (case-insensitive)", func(t *testing.T) {
		t.Setenv(varName, "YeS")
		assert.True(t, common.IsTestMode(plugin))
	})

	t.Run("returns false for 'false' value", func(t *testing.T) {
		t.Setenv(varName, "false")
		assert.False(t, common.IsTestMode(plugin))
	})

	t.Run("returns false for unset variable", func(t *testing.T) {
		os.Unsetenv(varName)
		assert.False(t, common.IsTestMode(plugin))
	})

	t.Run("returns false for random value", func(t *testing.T) {
		t.Setenv(varName, "maybe")
		assert.False(t, common.IsTestMode(plugin))
	})
}
