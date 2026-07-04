package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestProviderStepTypesMatchManifest(t *testing.T) {
	providerTypes := NewProvider().StepTypes()
	manifestTypes := readManifestStepTypes(t)

	if !slices.Equal(providerTypes, manifestTypes) {
		t.Fatalf("StepTypes() = %#v, manifest stepTypes = %#v", providerTypes, manifestTypes)
	}
}

func TestProviderBundleSummaryStep(t *testing.T) {
	root := writeTestBundle(t, t.TempDir(), "trace.ndjson")
	step, err := NewProvider().CreateStep(StepBundleSummary, "summary", map[string]any{
		"path": root,
	})
	if err != nil {
		t.Fatalf("CreateStep: %v", err)
	}

	result, err := step.Execute(t.Context(), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	output := result.Output
	if output["valid"] != true {
		t.Fatalf("valid = %v", output["valid"])
	}
	if output["run_id"] != "run-test" || output["status"] != "completed" {
		t.Fatalf("summary output = %#v", output)
	}
	if output["trace_count"] != 2 || output["step_count"] != 1 {
		t.Fatalf("summary counts = %#v", output)
	}
	if _, ok := output["trace"]; ok {
		t.Fatalf("summary output leaked trace content: %#v", output)
	}
}

func TestProviderFlowValidateStep(t *testing.T) {
	path := writeTestFlow(t, t.TempDir())
	step, err := NewProvider().CreateStep(StepFlowValidate, "flow", map[string]any{
		"path": path,
	})
	if err != nil {
		t.Fatalf("CreateStep: %v", err)
	}

	result, err := step.Execute(t.Context(), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["valid"] != true {
		t.Fatalf("valid = %v", result.Output["valid"])
	}
	if result.Output["start_at"] != "result" || result.Output["node_count"] != 1 {
		t.Fatalf("flow output = %#v", result.Output)
	}
}

func TestProviderStepRequiresPath(t *testing.T) {
	step, err := NewProvider().CreateStep(StepBundleValidate, "validate", nil)
	if err != nil {
		t.Fatalf("CreateStep: %v", err)
	}
	if _, err := step.Execute(t.Context(), nil, nil, nil, nil, nil); err == nil {
		t.Fatal("Execute succeeded, want missing path error")
	}
}

func readManifestStepTypes(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var manifest struct {
		Capabilities struct {
			StepTypes []string `json:"stepTypes"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}
	return manifest.Capabilities.StepTypes
}

func writeTestBundle(t *testing.T, root, tracePath string) string {
	t.Helper()
	now := time.Date(2026, 7, 4, 20, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	if err := os.MkdirAll(filepath.Join(root, "projections"), 0o755); err != nil {
		t.Fatalf("mkdir projections: %v", err)
	}
	writeTestJSON(t, filepath.Join(root, "manifest.json"), map[string]any{
		"schema":      "acpx.flow-run-bundle.v1",
		"runId":       "run-test",
		"flowName":    "test-flow",
		"startedAt":   now,
		"finishedAt":  now,
		"status":      "completed",
		"traceSchema": "acpx.flow-trace-event.v1",
		"paths": map[string]any{
			"flow":            "flow.json",
			"trace":           tracePath,
			"runProjection":   "projections/run.json",
			"liveProjection":  "projections/live.json",
			"stepsProjection": "projections/steps.json",
			"sessionsDir":     "sessions",
			"artifactsDir":    "artifacts",
		},
	})
	writeTestJSON(t, filepath.Join(root, "flow.json"), map[string]any{
		"schema":  "acpx.flow-definition-snapshot.v1",
		"name":    "test-flow",
		"startAt": "node-1",
		"nodes": map[string]any{
			"node-1": map[string]any{"nodeType": "compute"},
		},
		"edges": []any{},
	})
	writeTestJSON(t, filepath.Join(root, "projections", "run.json"), map[string]any{
		"runId":     "run-test",
		"flowName":  "test-flow",
		"startedAt": now,
		"updatedAt": now,
		"status":    "completed",
	})
	writeTestJSON(t, filepath.Join(root, "projections", "live.json"), map[string]any{
		"runId":     "run-test",
		"flowName":  "test-flow",
		"startedAt": now,
		"updatedAt": now,
		"status":    "completed",
	})
	writeTestJSON(t, filepath.Join(root, "projections", "steps.json"), []map[string]any{{
		"attemptId":  "node-1#1",
		"nodeId":     "node-1",
		"nodeType":   "compute",
		"outcome":    "ok",
		"startedAt":  now,
		"finishedAt": now,
	}})
	trace := `{"seq":1,"at":"` + now + `","scope":"node","type":"node_started","runId":"run-test","nodeId":"node-1","attemptId":"node-1#1","payload":{}}` + "\n" +
		`{"seq":2,"at":"` + now + `","scope":"node","type":"node_outcome","runId":"run-test","nodeId":"node-1","attemptId":"node-1#1","payload":{"outcome":"ok"}}` + "\n"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(tracePath)), []byte(trace), 0o600); err != nil {
		t.Fatalf("write trace: %v", err)
	}
	return root
}

func writeTestFlow(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "flow.json")
	writeTestJSON(t, path, map[string]any{
		"format_version": 1,
		"start_at":       "result",
		"nodes": []map[string]any{
			{"id": "result", "type": "compute", "value": map[string]any{"ok": true}},
		},
	})
	return path
}

func writeTestJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
