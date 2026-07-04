package internal

import "github.com/GoCodeAlone/workflow/plugin/external/sdk"

// Version is injected by release builds.
var Version = "dev"

// Provider implements Workflow external plugin provider interfaces.
type Provider struct{}

// NewProvider creates a Workflow ACPX plugin provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Manifest implements sdk.PluginProvider.
func (p *Provider) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-acpx",
		Version:     sdk.ResolveBuildVersion(Version),
		Author:      "GoCodeAlone",
		Description: "ACPX durable bundle validation and summaries",
	}
}

// StepTypes implements sdk.StepProvider.
func (p *Provider) StepTypes() []string {
	return stepTypes()
}

// CreateStep implements sdk.StepProvider.
func (p *Provider) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	if !isStepType(typeName) {
		return nil, unknownStepType(typeName)
	}
	return &sdkStep{stepType: typeName, name: name, config: cloneConfig(config)}, nil
}
