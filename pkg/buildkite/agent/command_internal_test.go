package agent

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := &config{
			command: func(_ string, _ ...string) *exec.Cmd {
				cmd := exec.Command("echo", "hello world")
				return cmd
			},
		}
		output, err := cfg.runCommand(t.Context(), "echo", "hello world")
		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Contains(t, *output, "hello world")
	})

	t.Run("failure", func(t *testing.T) {
		cfg := &config{
			command: func(_ string, _ ...string) *exec.Cmd {
				// This command will fail
				cmd := exec.Command("false")
				return cmd
			},
		}
		output, err := cfg.runCommand(t.Context(), "false")
		require.Error(t, err)
		assert.Nil(t, output)
	})
}
