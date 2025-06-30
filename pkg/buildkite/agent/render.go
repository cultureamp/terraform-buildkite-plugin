package agent

import (
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
)

// renderTemplate parses and applies a template file with the provided data.
func (a *config) renderTemplate(templatePath string, data any) (string, error) {
	log.Info().Str("template", templatePath).Msg("Parsing template")
	// Parse the template file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}
	log.Info().Msg("Applying template")
	// Execute the template with the provided data
	var rendered strings.Builder
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return "", err
	}
	log.Info().Msg("Template applied successfully")
	return rendered.String(), nil
}
