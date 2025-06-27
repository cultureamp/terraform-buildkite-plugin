package handler

import (
	"github.com/spf13/cobra"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/schema"
)

// MockHandler is a test double for Handler.
type MockHandler struct {
	HandleFunc      func(schema.Config, *HandleOptions) func(cmd *cobra.Command, args []string) error
	HandleReturnErr error
}

func (m *MockHandler) Handle(s schema.Config, opts *HandleOptions) func(cmd *cobra.Command, args []string) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(s, opts)
	}
	return func(_ *cobra.Command, _ []string) error {
		return m.HandleReturnErr
	}
}
