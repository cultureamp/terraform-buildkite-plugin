package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	t.Run("renders template with data", func(t *testing.T) {
		tmplContent := "Hello, {{.Name}}!"
		dir := t.TempDir()
		tmplPath := filepath.Join(dir, "test.tmpl")
		err := os.WriteFile(tmplPath, []byte(tmplContent), 0644)
		require.NoError(t, err)

		cfg := &config{}
		data := map[string]string{"Name": "World"}
		result, err := cfg.renderTemplate(tmplPath, data)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", strings.TrimSpace(result))
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		cfg := &config{}
		_, err := cfg.renderTemplate("/nonexistent/file.tmpl", nil)
		require.Error(t, err)
	})

	t.Run("returns error for invalid template", func(t *testing.T) {
		dir := t.TempDir()
		tmplPath := filepath.Join(dir, "bad.tmpl")
		err := os.WriteFile(tmplPath, []byte("{{.Name"), 0644)
		require.NoError(t, err)
		cfg := &config{}
		_, err = cfg.renderTemplate(tmplPath, map[string]string{"Name": "fail"})
		require.Error(t, err)
	})
}
