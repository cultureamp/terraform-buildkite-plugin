package handler

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

// FileExtensionValidator checks if a file has one of the allowed extensions.
func FileExtensionValidator(fl validator.FieldLevel) bool {
	// Ensure the field is of type string
	if fl.Field().Kind() != reflect.String {
		return false
	}

	// Trim leading and trailing whitespace
	fileName := strings.TrimSpace(fl.Field().String())

	// If the string is empty, fail the validation
	if fileName == "" {
		return false
	}

	// Extract the file extension and convert it to lowercase
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	if ext == "" {
		return false
	}

	// Get the list of allowed extensions from the tag and split them by comma
	allowedExts := fl.Param()
	allowedList := strings.Split(allowedExts, " ")

	// Check if the file extension is in the allowed list (case-insensitive)
	for _, allowed := range allowedList {
		allowed = strings.ToLower(strings.TrimSpace(allowed))
		if ext == allowed {
			return true
		}
	}

	return false
}

// ValidateOptions validates the CLI options for schema generation.
// It checks required fields and file extension constraints.
func ValidateOptions(opts *HandleOptions) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// Only register the extension validator once per validator instance.
	if err := validate.RegisterValidation("extension", FileExtensionValidator); err != nil {
		return fmt.Errorf("failed to register validation: %s", "extension")
	}
	if err := validate.Struct(opts); err != nil {
		return fmt.Errorf("failed to validate options: %w", err)
	}
	return nil
}
