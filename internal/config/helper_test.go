package config_test

import (
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/buildkite/group"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	//nolint:reassign // sinencing the global logger to avoid output during tests
	log.Logger = zerolog.New(nil)
	group.SetOutput(nil)
	m.Run()
}
