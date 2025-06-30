package agent_test

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	//nolint:reassign // sinencing the global logger to avoid output during tests
	log.Logger = zerolog.New(nil)
	m.Run()
}
