package internal

import (
	"testing"

	"github.com/GoCodeAlone/workflow/capability"
	"github.com/GoCodeAlone/workflow/interfaces"
	"github.com/GoCodeAlone/workflow/plugin"
	"github.com/GoCodeAlone/workflow/schema"
)

func TestEnginePluginRunsBundleSummaryStep(t *testing.T) {
	root := writeTestBundle(t, t.TempDir(), "trace.ndjson")
	loader := plugin.NewPluginLoader(capability.NewRegistry(), schema.NewModuleSchemaRegistry())
	if err := loader.LoadPlugin(NewEnginePlugin()); err != nil {
		t.Fatalf("LoadPlugin: %v", err)
	}
	factory := loader.StepFactories()[StepBundleSummary]
	if factory == nil {
		t.Fatalf("missing step factory for %s", StepBundleSummary)
	}
	step, err := factory("summary", map[string]any{"path": root}, nil)
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	pipelineStep, ok := step.(interfaces.PipelineStep)
	if !ok {
		t.Fatalf("factory returned %T, want interfaces.PipelineStep", step)
	}

	result, err := pipelineStep.Execute(t.Context(), interfaces.NewPipelineContext(nil, nil))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	output := result.Output
	if output["valid"] != true {
		t.Fatalf("valid = %v", output["valid"])
	}
	if output["run_id"] != "run-test" || output["trace_count"] != 2 {
		t.Fatalf("summary output = %#v", output)
	}
}
