// Package outputs provides adapters for integrating existing components
// with the orchestrator interfaces.
package outputs

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/xphir/terraform-buildkite-plugin/pkg/buildkite/agent"
)

type buildkiteAnnotatorConfig struct {
	agent  agent.Agent
	config *BuildkiteAnnotation
}

// BuildkiteAnnotatorOptions allows functional options for customizing config.
type BuildkiteAnnotatorOptions func(*buildkiteAnnotatorConfig)

// WithAgent allows injecting a custom CommandFn (e.g., for testing).
func WithAgent(a agent.Agent) BuildkiteAnnotatorOptions {
	return func(r *buildkiteAnnotatorConfig) {
		if a != nil {
			r.agent = a
		}
	}
}

// WithConfig allows setting a custom BuildkiteAnnotation configuration.
func WithConfig(c *BuildkiteAnnotation) BuildkiteAnnotatorOptions {
	return func(r *buildkiteAnnotatorConfig) {
		if c != nil {
			r.config = c
		}
	}
}

// NewBuildkiteAnnotator creates a new annotator adapter for Buildkite annotations.
func NewBuildkiteAnnotator(opts ...BuildkiteAnnotatorOptions) Outputer {
	outputer := &buildkiteAnnotatorConfig{
		agent: agent.NewAgent(),
	}
	for _, opt := range opts {
		opt(outputer)
	}
	return outputer
}

// Ouput creates a success annotation for completed operations.
func (a *buildkiteAnnotatorConfig) Ouput(ctx context.Context, _ *tfjson.Plan, stage Stage, data any) error {
	_, err := a.agent.AnnotateWithTemplate(ctx, a.config.Template, data,
		agent.WithAppend(false),
		agent.WithStyle(stage.toBuildkiteAnnotationStyle()),
		agent.WithContext(a.config.Context),
	)
	if err != nil {
		return fmt.Errorf("failed to create Buildkite annotation: %w", err)
	}
	return nil
}

// toBuildkiteAnnotationStyle converts the Stage to a Buildkite annotation style.
func (s Stage) toBuildkiteAnnotationStyle() agent.AnnotationStyle {
	switch s {
	case PlanFailure, ApplyFailure, ValidationFailure, UnexpectedFailure:
		return agent.StyleError
	case PlanSuccessWithChanges, ValidationSuccess, ApplySuccess:
		return agent.StyleSuccess
	case PlanSuccessNoChanges:
		return agent.StyleInfo
	default:
		return agent.StyleInfo
	}
}
