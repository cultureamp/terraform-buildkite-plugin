package handler_test

import (
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/handler"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestFileExtensionValidator(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("extension", handler.FileExtensionValidator)
	require.NoError(t, err)

	type Config struct {
		ConfigFile string `validate:"required,extension=yaml yml json txt"`
	}

	type NonStringConfig struct {
		InvalidField int `validate:"extension=yaml yml"` // Non-string field for testing
	}

	testCases := []struct {
		name       string
		cfg        any
		shouldPass bool
	}{
		// Valid cases
		{name: "Valid YAML file", cfg: Config{ConfigFile: "config.yaml"}, shouldPass: true},
		{name: "Valid YML file", cfg: Config{ConfigFile: "config.yml"}, shouldPass: true},
		{name: "Valid JSON file", cfg: Config{ConfigFile: "config.json"}, shouldPass: true},
		{name: "Valid TXT file", cfg: Config{ConfigFile: "config.txt"}, shouldPass: true},
		{name: "Uppercase extension", cfg: Config{ConfigFile: "CONFIG.YAML"}, shouldPass: true},
		{name: "Mixed case extension", cfg: Config{ConfigFile: "config.YaMl"}, shouldPass: true},
		{name: "Trailing newline", cfg: Config{ConfigFile: "config.yaml\n"}, shouldPass: true},
		{name: "Trailing spaces", cfg: Config{ConfigFile: "config.yaml   "}, shouldPass: true},
		{name: "Multiple dots in file name", cfg: Config{ConfigFile: "archive.tar.yaml"}, shouldPass: true},
		{name: "Hidden file with valid extension", cfg: Config{ConfigFile: ".config.yaml"}, shouldPass: true},

		// Invalid cases
		{name: "Invalid extension", cfg: Config{ConfigFile: "config.exe"}, shouldPass: false},
		{name: "No extension", cfg: Config{ConfigFile: "config"}, shouldPass: false},
		{name: "Empty file name", cfg: Config{ConfigFile: ""}, shouldPass: false},
		{name: "Whitespace file name", cfg: Config{ConfigFile: "   "}, shouldPass: false},
		{name: "Invalid nested extension", cfg: Config{ConfigFile: "archive.tar.exe"}, shouldPass: false},
		{name: "Hidden file without extension", cfg: Config{ConfigFile: ".config"}, shouldPass: false},
		{name: "File with only dot", cfg: Config{ConfigFile: "."}, shouldPass: false},
		{name: "File with dot and spaces", cfg: Config{ConfigFile: ".   "}, shouldPass: false},

		// Non-string field cases
		{name: "Non-string field", cfg: NonStringConfig{InvalidField: 12345}, shouldPass: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertValidation(t, validate, tc.cfg, tc.shouldPass)
		})
	}
}

func assertValidation(t *testing.T, validate *validator.Validate, cfg interface{}, shouldPass bool) {
	err := validate.Struct(cfg)
	if shouldPass && err != nil {
		t.Errorf("Expected validation to pass for '%v', but got error: %v", cfg, err)
	}
	if !shouldPass && err == nil {
		t.Errorf("Expected validation to fail for '%v', but it passed", cfg)
	}
}

func TestValidateOptions(t *testing.T) {
	type testCase struct {
		name      string
		opts      handler.HandleOptions
		expectErr bool
		errMsg    string
	}

	tests := []testCase{
		{
			name:      "Valid yaml extension",
			opts:      handler.HandleOptions{OutputFile: "foo.yaml"},
			expectErr: false,
		},
		{
			name:      "Valid yml extension",
			opts:      handler.HandleOptions{OutputFile: "foo.yml"},
			expectErr: false,
		},
		{
			name:      "Missing output file",
			opts:      handler.HandleOptions{OutputFile: ""},
			expectErr: true,
			errMsg:    "failed to validate options",
		},
		{
			name:      "Invalid extension",
			opts:      handler.HandleOptions{OutputFile: "foo.txt"},
			expectErr: true,
			errMsg:    "failed to validate options",
		},
		{
			name:      "Whitespace only",
			opts:      handler.HandleOptions{OutputFile: "   "},
			expectErr: true,
			errMsg:    "failed to validate options",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handler.ValidateOptions(&tc.opts)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
