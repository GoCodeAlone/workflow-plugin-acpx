package internal

import (
	"context"
	"fmt"

	"github.com/GoCodeAlone/modular"
	"github.com/GoCodeAlone/workflow/interfaces"
	"github.com/GoCodeAlone/workflow/plugin"
)

type enginePlugin struct {
	plugin.BaseEnginePlugin
}

// NewEnginePlugin creates an in-process Workflow engine plugin for tests and
// embedded hosts that do not need the external plugin process boundary.
func NewEnginePlugin() plugin.EnginePlugin {
	version := enginePluginVersion()
	return &enginePlugin{
		BaseEnginePlugin: plugin.BaseEnginePlugin{
			BaseNativePlugin: plugin.BaseNativePlugin{
				PluginName:        "workflow-plugin-acpx",
				PluginVersion:     version,
				PluginDescription: "ACPX durable bundle validation and summaries",
			},
			Manifest: plugin.PluginManifest{
				Name:        "workflow-plugin-acpx",
				Version:     version,
				Author:      "GoCodeAlone",
				Description: "ACPX durable bundle validation and summaries",
				Tier:        plugin.TierCommunity,
				StepTypes:   stepTypes(),
			},
		},
	}
}

func enginePluginVersion() string {
	if Version == "" || Version == "dev" || Version == "(devel)" {
		return "0.0.0"
	}
	return Version
}

func (p *enginePlugin) StepFactories() map[string]plugin.StepFactory {
	factories := make(map[string]plugin.StepFactory, len(stepTypes()))
	for _, stepType := range stepTypes() {
		stepType := stepType
		factories[stepType] = func(name string, cfg map[string]any, _ modular.Application) (any, error) {
			return &pipelineStep{stepType: stepType, name: name, config: cloneConfig(cfg)}, nil
		}
	}
	return factories
}

type pipelineStep struct {
	stepType string
	name     string
	config   map[string]any
}

func (s *pipelineStep) Name() string {
	return s.name
}

func (s *pipelineStep) Execute(ctx context.Context, _ *interfaces.PipelineContext) (*interfaces.StepResult, error) {
	output, err := executeStep(ctx, s.stepType, s.config)
	if err != nil {
		return nil, fmt.Errorf("%s step %q: %w", s.stepType, s.name, err)
	}
	return &interfaces.StepResult{Output: output}, nil
}
