package outputs

import (
	"context"

	tfjson "github.com/hashicorp/terraform-json"
)

type Stage string

const (
	PlanFailure            Stage = "plan_failure"
	ApplyFailure           Stage = "apply_failure"
	ValidationFailure      Stage = "validation_failure"
	UnexpectedFailure      Stage = "unexpected_failure"
	PlanSuccessNoChanges   Stage = "plan_success_no_changes"
	PlanSuccessWithChanges Stage = "plan_success_with_changes"
	ValidationSuccess      Stage = "validation_success"
	ApplySuccess           Stage = "apply_success"
)

type Outputer interface {
	Ouput(ctx context.Context, plan *tfjson.Plan, stage Stage, data any) error
}
