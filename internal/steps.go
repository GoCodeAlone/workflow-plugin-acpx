package internal

import (
	"context"
	"fmt"
	"maps"

	"github.com/GoCodeAlone/workflow-plugin-acpx/acpxruntime"
	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

const (
	StepBundleValidate = "acpx.bundle_validate"
	StepBundleSummary  = "acpx.bundle_summary"
	StepFlowValidate   = "acpx.flow_validate"
)

type sdkStep struct {
	stepType string
	name     string
	config   map[string]any
}

func (s *sdkStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, _ map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	output, err := executeStep(ctx, s.stepType, mergeConfig(s.config, config))
	if err != nil {
		return nil, fmt.Errorf("%s step %q: %w", s.stepType, s.name, err)
	}
	return &sdk.StepResult{Output: output}, nil
}

func executeStep(ctx context.Context, stepType string, config map[string]any) (map[string]any, error) {
	path, err := configPath(config)
	if err != nil {
		return nil, err
	}
	switch stepType {
	case StepBundleValidate:
		if err := acpxruntime.ValidateBundle(ctx, path); err != nil {
			return nil, err
		}
		return map[string]any{
			"valid": true,
			"path":  path,
		}, nil
	case StepBundleSummary:
		summary, err := acpxruntime.ReplaySummary(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"valid":         true,
			"path":          path,
			"run_id":        summary.RunID,
			"flow_name":     summary.FlowName,
			"status":        summary.Status,
			"trace_count":   summary.TraceCount,
			"step_count":    summary.StepCount,
			"session_count": summary.SessionCount,
		}, nil
	case StepFlowValidate:
		summary, err := acpxruntime.ValidateFlowFile(path)
		if err != nil {
			return nil, err
		}
		output := map[string]any{
			"valid":      true,
			"path":       path,
			"start_at":   summary.StartAt,
			"node_count": summary.NodeCount,
		}
		if summary.Name != "" {
			output["name"] = summary.Name
		}
		if len(summary.Requires) > 0 {
			output["requires"] = summary.Requires
		}
		return output, nil
	default:
		return nil, unknownStepType(stepType)
	}
}

func stepTypes() []string {
	return []string{StepBundleValidate, StepBundleSummary, StepFlowValidate}
}

func isStepType(stepType string) bool {
	for _, candidate := range stepTypes() {
		if stepType == candidate {
			return true
		}
	}
	return false
}

func unknownStepType(stepType string) error {
	return fmt.Errorf("unknown step type: %s", stepType)
}

func configPath(config map[string]any) (string, error) {
	if config == nil {
		return "", fmt.Errorf("path is required")
	}
	path, _ := config["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	return path, nil
}

func cloneConfig(config map[string]any) map[string]any {
	if config == nil {
		return nil
	}
	clone := make(map[string]any, len(config))
	maps.Copy(clone, config)
	return clone
}

func mergeConfig(base, override map[string]any) map[string]any {
	merged := cloneConfig(base)
	if merged == nil {
		merged = map[string]any{}
	}
	maps.Copy(merged, override)
	return merged
}
