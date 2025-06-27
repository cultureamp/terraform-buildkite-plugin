package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// CommandFn is a function type for creating exec.Cmd, allowing DI for testing.
type CommandFn func(command string, args ...string) *exec.Cmd

// runCommand executes a command with the provided arguments and returns its output.
func (c *config) runCommand(_ context.Context, command string, args ...string) (*string, error) {
	cmd := c.command(command, args...)
	log.Debug().Str("command", command).Strs("args", args).Msg("Executing command")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Error().
			Str("command", command).
			Strs("args", args).
			Str("stderr", stderr.String()).
			Err(err).
			Msg("Command execution failed")
		return nil, fmt.Errorf("command `%s` failed: %w: %s", command, err, stderr.String())
	}
	output := out.String()
	log.Debug().Str("command", command).Strs("args", args).Str("stdout", output).Msg("Command executed successfully")
	return &output, nil
}
