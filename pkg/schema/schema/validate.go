package schema

import (
	"errors"
	"fmt"
	"reflect"
)

func (c *config) Validate() error {
	var errs []error
	if err := validateInputSchema(c.Schema); err != nil {
		errs = append(errs, err)
	}
	if err := validateProperties(c.Properties); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	// Aggregate errors into a single error
	msg := "validation errors:"
	for _, err := range errs {
		msg += "\n - " + err.Error()
	}
	return errors.New(msg)
}

// validateInputSchema ensures the input schema is a non-nil struct or a pointer to a struct.
func validateInputSchema(input any) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	t := reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference pointer types
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %T", input)
	}

	return nil
}

// validateProperties ensures the GeneratorConfig.Properties is not nil.
func validateProperties(properties *PluginProperties) error {
	if properties == nil {
		return errors.New("properties cannot be nil")
	}
	return nil
}
