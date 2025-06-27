package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/xphir/terraform-buildkite-plugin/internal/common"
	i "github.com/xphir/terraform-buildkite-plugin/internal/plugin/initiator"
	o "github.com/xphir/terraform-buildkite-plugin/internal/plugin/orchestrator"
	a "github.com/xphir/terraform-buildkite-plugin/pkg/buildkite/agent"
)

type ExitStatus int

const (
	Success              ExitStatus = 0
	UnexpectedFailure    ExitStatus = 1
	HandledFailure       ExitStatus = 2
	NoWorkingDirectories ExitStatus = 3
	TestModeEarlyExit    ExitStatus = 10
)

// GetName returns the string representation of an ExitStatus.
func (e ExitStatus) GetName() string {
	switch e {
	case Success:
		return "Success"
	case UnexpectedFailure:
		return "UnexpectedFailure"
	case HandledFailure:
		return "HandledFailure"
	case NoWorkingDirectories:
		return "NoWorkingDirectories"
	case TestModeEarlyExit:
		return "TestModeEarlyExit"
	default:
		return fmt.Sprintf("Unknown(%d)", int(e))
	}
}

// ToInt converts an ExitStatus to its integer representation.
func (e ExitStatus) ToInt() int {
	return int(e)
}

type Context struct {
	Name    string // Name of the plugin
	Version string // Version of the plugin
}

type Handler interface {
	Handle(ctx context.Context, context *Context) (ExitStatus, error)
}

type handlerConfig struct {
	tExecPath       string            // Path to the Terraform executable
	agent           a.Agent           // Buildkite agent for uploading pipelines and annotations
	pluginInitiator i.PluginInitiator // The initiator interface
}

type HandlerOption func(*handlerConfig)

func WithTerraformExecPath(path string) HandlerOption {
	return func(h *handlerConfig) {
		if path != "" {
			h.tExecPath = path
		}
	}
}

func WithAgentInterface(a a.Agent) HandlerOption {
	return func(h *handlerConfig) {
		if a != nil {
			h.agent = a
		}
	}
}

func WithInitatorInterface(i i.PluginInitiator) HandlerOption {
	return func(h *handlerConfig) {
		if i != nil {
			h.pluginInitiator = i
		}
	}
}

// NewHandler creates a new instance of the plugin with the provided configuration options.
func NewHandler(
	opts ...HandlerOption,
) Handler {
	defaults := &handlerConfig{
		tExecPath:       "", // Default to empty, will auto-discover "terraform" on PATH
		agent:           a.NewAgent(),
		pluginInitiator: i.NewInitiator(),
	}
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}

func (h *handlerConfig) Handle(
	ctx context.Context,
	context *Context,
) (ExitStatus, error) {
	payload, err := h.pluginInitiator.ParsePlugin(ctx, context.Name)
	if err != nil {
		return UnexpectedFailure, err
	}
	if common.IsTestMode(context.Name) {
		if writeErr := common.WritePrettyJSON(payload.Plugin, os.Stderr); writeErr != nil {
			log.Warn().Err(writeErr).Msg("failed to pretty print plugin config")
		}
		log.Info().Msg("test mode is enabled, skipping plugin execution")
		return TestModeEarlyExit, nil
	}
	if len(payload.WorkingDirectories) == 0 {
		log.Warn().Msg("no working directories specified, skipping plugin execution")
		return NoWorkingDirectories, nil
	}
	log.Info().Int("workspaces", len(payload.WorkingDirectories)).Msg("starting plugin execution across workspaces")
	log.Debug().Msg("creating orchestrator for plugin execution")
	orchestrator, err := o.NewOrchestrator(
		payload.Plugin,
		payload.Validators,
		payload.Outputers,
		o.WithAgentInterface(h.agent),
		o.WithTerraformExecPath(h.tExecPath),
	)
	if err != nil {
		return UnexpectedFailure, err
	}
	failures := []o.WorkspaceResult{}
	for _, workingDir := range payload.WorkingDirectories {
		workdirName := filepath.Base(workingDir)
		log.Info().Str("workspace", workdirName).
			Msg("running orchestrator for workspace")
		result := orchestrator.Run(ctx, workingDir)
		if result != nil && !result.Success {
			log.Warn().Str("workspace", workdirName).Msg("workspace execution failed")
			failures = append(failures, *orchestrator.Run(ctx, workingDir))
		} else {
			log.Info().Str("workspace", workdirName).
				Msg("workspace execution succeeded")
		}
	}

	if len(failures) > 0 {
		log.Error().Int("failures", len(failures)).Msg("plugin execution failed in some workspaces")
		for _, failure := range failures {
			log.Error().Interface("workspace", failure).Msg("workspace execution failure")
		}
		return HandledFailure, nil
	}
	log.Info().Msg("plugin execution completed successfully across all workspaces")
	return Success, nil
}
