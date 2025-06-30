package agent_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/buildkite/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_New(t *testing.T) {
	ag := agent.NewAgent()
	assert.NotNil(t, ag)
}

func TestAgent_UploadPipeline(t *testing.T) {
	t.Run("calls runCommand", func(t *testing.T) {
		called := false
		agentWithMock := agent.NewAgent(agent.WithCommandFn(func(_ string, _ ...string) *exec.Cmd {
			called = true
			return exec.Command("echo", "pipeline uploaded")
		}))
		result, err := agentWithMock.UploadPipeline(t.Context(), "pipeline.yml")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, called)
	})
}

func TestAgent_Annotate(t *testing.T) {
	t.Run("calls runCommand", func(t *testing.T) {
		called := false
		agentWithMock := agent.NewAgent(agent.WithCommandFn(func(_ string, _ ...string) *exec.Cmd {
			called = true
			return exec.Command("echo", "annotated")
		}))
		result, err := agentWithMock.Annotate(t.Context())
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, called)
	})
}

func TestAgent_AnnotateWithTemplate(t *testing.T) {
	t.Run("renders and annotates", func(t *testing.T) {
		didAnnotate := false
		agentWithMock := agent.NewAgent(agent.WithCommandFn(func(_ string, _ ...string) *exec.Cmd {
			didAnnotate = true
			return exec.Command("echo", "annotated with template")
		}))
		// Write a temp template file
		tmpl := "Hello, {{.Name}}!"
		tempFile := t.TempDir() + "/tmpl.txt"
		err := os.WriteFile(tempFile, []byte(tmpl), 0644)
		require.NoError(t, err)
		result, err := agentWithMock.AnnotateWithTemplate(t.Context(), tempFile, map[string]string{"Name": "Test"})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, didAnnotate)
	})

	t.Run("render error", func(t *testing.T) {
		agentWithMock := agent.NewAgent(agent.WithCommandFn(func(_ string, _ ...string) *exec.Cmd {
			return exec.Command("echo", "should not be called")
		}))
		_, err := agentWithMock.AnnotateWithTemplate(t.Context(), "/nonexistent/file", nil)
		require.Error(t, err)
	})
}
