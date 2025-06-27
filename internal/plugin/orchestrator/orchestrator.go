package orchestrator

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	o "github.com/xphir/terraform-buildkite-plugin/internal/adapters/outputs"
	v "github.com/xphir/terraform-buildkite-plugin/internal/adapters/validators"
	c "github.com/xphir/terraform-buildkite-plugin/internal/config"
	a "github.com/xphir/terraform-buildkite-plugin/pkg/buildkite/agent"
)

type WorkspaceResult struct {
	Success    bool
	Stage      string
	WorkingDir string
	Error      interface{}
}

type PluginOrchestrator interface {
	Plan(ctx context.Context, workingDir string) *WorkspaceResult
	Apply(ctx context.Context, workingDir string) *WorkspaceResult
	Run(ctx context.Context, workingDir string) *WorkspaceResult
}

type orchestratorConfig struct {
	tExecPath  string
	agent      a.Agent
	plugin     *c.Plugin
	validators []v.Validator
	outputers  []o.Outputer
}

type Option func(*orchestratorConfig)

func WithTerraformExecPath(path string) Option {
	return func(h *orchestratorConfig) {
		if path != "" {
			h.tExecPath = path
		}
	}
}

func WithAgentInterface(a a.Agent) Option {
	return func(h *orchestratorConfig) {
		if a != nil {
			h.agent = a
		}
	}
}

// NewOrchestrator creates a new instance of the plugin with the provided configuration options.
func NewOrchestrator(
	plugin *c.Plugin,
	validators []v.Validator,
	outputers []o.Outputer,
	opts ...Option,
) (PluginOrchestrator, error) {
	tExecPath := ""
	if plugin.Terraform != nil && plugin.Terraform.ExecPath != nil {
		tExecPath = *plugin.Terraform.ExecPath
	}
	defaults := &orchestratorConfig{
		tExecPath:  tExecPath,
		agent:      a.NewAgent(),
		plugin:     plugin,
		validators: validators,
		outputers:  outputers,
	}
	for _, opt := range opts {
		opt(defaults)
	}
	if defaults.tExecPath == "" {
		p, err := exec.LookPath("terraform")
		if err != nil {
			return nil, fmt.Errorf("terraform binary not found in PATH: %w", err)
		}
		defaults.tExecPath = p
	}
	return defaults, nil
}

func (o *orchestratorConfig) Run(
	ctx context.Context,
	workingDir string,
) *WorkspaceResult {
	switch o.plugin.Mode {
	case c.Plan:
		return o.Plan(ctx, workingDir)
	case c.Apply:
		return o.Apply(ctx, workingDir)
	default:
		return &WorkspaceResult{
			Success:    false,
			Stage:      "validation",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("unsupported plugin mode: %s", o.plugin.Mode),
		}
	}
}

func (o *orchestratorConfig) Plan(ctx context.Context, workingDir string) *WorkspaceResult {
	planFile := path.Join(workingDir, "plan.binary")
	tf, result := o.initSteps(ctx, workingDir)
	if result != nil {
		return result
	}
	planJSON, result := o.planSteps(ctx, tf, planFile, workingDir)
	if result != nil {
		return result
	}
	result = o.validateSteps(ctx, planJSON, workingDir)
	if result != nil {
		return result
	}
	return &WorkspaceResult{
		Success:    true,
		Stage:      "planning",
		WorkingDir: workingDir,
		Error:      nil,
	}
}

func (o *orchestratorConfig) Apply(ctx context.Context, workingDir string) *WorkspaceResult {
	planFile := path.Join(workingDir, "plan.binary")
	tf, result := o.initSteps(ctx, workingDir)
	if result != nil {
		return result
	}
	planJSON, result := o.planSteps(ctx, tf, planFile, workingDir)
	if result != nil {
		return result
	}
	result = o.validateSteps(ctx, planJSON, workingDir)
	if result != nil {
		return result
	}
	if err := tf.Apply(ctx, tfexec.DirOrPlan(planFile)); err != nil {
		return &WorkspaceResult{
			Success:    false,
			Stage:      "applying",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("failed to apply Terraform plan: %v", err),
		}
	}
	return &WorkspaceResult{
		Success:    true,
		Stage:      "apply",
		WorkingDir: workingDir,
		Error:      nil,
	}
}

func (o *orchestratorConfig) newTerraform(workingDir string) (*tfexec.Terraform, error) {
	tf, err := tfexec.NewTerraform(workingDir, o.tExecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform runner: %w", err)
	}
	return tf, nil
}

func (o *orchestratorConfig) initSteps(ctx context.Context, workingDir string) (*tfexec.Terraform, *WorkspaceResult) {
	tf, err := o.newTerraform(workingDir)
	if err != nil {
		return nil, &WorkspaceResult{
			Success:    false,
			Stage:      "initialization",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("failed to initialize Terraform: %v", err),
		}
	}
	var initOpts []tfexec.InitOption
	if ti := o.plugin.Terraform; ti != nil && ti.InitOptions != nil {
		opts := ti.InitOptions
		if opts.Get != nil {
			initOpts = append(initOpts, tfexec.Get(*opts.Get))
		}
		if opts.PluginDir != nil {
			initOpts = append(initOpts, tfexec.PluginDir(*opts.PluginDir))
		}
	}
	if err = tf.Init(ctx, initOpts...); err != nil {
		return nil, &WorkspaceResult{
			Success:    false,
			Stage:      "initialization",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("failed to run terraform init: %v", err),
		}
	}
	return tf, nil
}

func (o *orchestratorConfig) planSteps(
	ctx context.Context,
	tf *tfexec.Terraform,
	planFile string,
	workingDir string,
) (*tfjson.Plan, *WorkspaceResult) {
	hasChanges, err := tf.Plan(ctx, tfexec.Out(planFile))
	if err != nil {
		return nil, &WorkspaceResult{
			Success:    false,
			Stage:      "planning",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("failed to run terraform plan: %v", err),
		}
	}
	if !hasChanges {
		return nil, &WorkspaceResult{
			Success:    true,
			Stage:      "planning",
			WorkingDir: workingDir,
			Error:      "no changes detected in the Terraform plan",
		}
	}
	plan, err := tf.ShowPlanFile(ctx, planFile)
	if err != nil {
		return nil, &WorkspaceResult{
			Success:    false,
			Stage:      "showing plan",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("failed to show plan file: %v", err),
		}
	}
	return plan, nil
}

func (o *orchestratorConfig) validateSteps(
	ctx context.Context,
	plan *tfjson.Plan,
	workingDir string,
) *WorkspaceResult {
	var validationFalures []v.ValidationResult
	for _, validator := range o.validators {
		result, err := validator.Validate(ctx, plan)
		if err != nil {
			return &WorkspaceResult{
				Success:    false,
				Stage:      "validation",
				WorkingDir: workingDir,
				Error:      fmt.Sprintf("validation failed: %v", err),
			}
		}
		if !result.Passed {
			validationFalures = append(validationFalures, result)
		}
	}
	if len(validationFalures) > 0 {
		return &WorkspaceResult{
			Success:    false,
			Stage:      "validation",
			WorkingDir: workingDir,
			Error:      fmt.Sprintf("validation failed with %d issues", len(validationFalures)),
		}
	}
	return nil
}
