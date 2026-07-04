package acpxruntime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRuntimeReplaySummaryLoadsBundle(t *testing.T) {
	root := writeMinimalBundle(t, t.TempDir(), "trace.ndjson")

	summary, err := ReplaySummary(t.Context(), root)
	if err != nil {
		t.Fatalf("ReplaySummary: %v", err)
	}
	if summary.RunID != "run-test" {
		t.Fatalf("RunID = %q", summary.RunID)
	}
	if summary.Status != "completed" {
		t.Fatalf("Status = %q", summary.Status)
	}
	if summary.TraceCount != 2 || summary.StepCount != 1 {
		t.Fatalf("summary counts = trace %d steps %d", summary.TraceCount, summary.StepCount)
	}
}

func TestRuntimeValidateBundleRejectsEscapingPath(t *testing.T) {
	root := writeMinimalBundle(t, t.TempDir(), "../trace.ndjson")

	err := ValidateBundle(t.Context(), root)
	if err == nil {
		t.Fatal("ValidateBundle succeeded, want path escape error")
	}
	if !strings.Contains(err.Error(), "bundle path") {
		t.Fatalf("error = %v, want bundle path validation", err)
	}
}

func TestRuntimeValidateFlowDefinition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flow.json")
	writeJSON(t, path, map[string]any{
		"format_version": 1,
		"start_at":       "result",
		"nodes": []map[string]any{
			{"id": "result", "type": "compute", "value": map[string]any{"ok": true}},
		},
	})

	summary, err := ValidateFlowFile(path)
	if err != nil {
		t.Fatalf("ValidateFlowFile: %v", err)
	}
	if summary.NodeCount != 1 || summary.StartAt != "result" {
		t.Fatalf("summary = %#v", summary)
	}
}

func TestRuntimeValidateFlowDefinitionRejectsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flow.json")
	writeJSON(t, path, map[string]any{
		"format_version": 1,
		"start_at":       "missing",
		"nodes":          []map[string]any{{"id": "result", "type": "compute", "value": "ok"}},
	})

	if _, err := ValidateFlowFile(path); err == nil {
		t.Fatal("ValidateFlowFile succeeded, want invalid definition error")
	}
}

func TestRuntimeValidateBundleHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := ValidateBundle(ctx, t.TempDir()); err == nil {
		t.Fatal("ValidateBundle succeeded with cancelled context")
	}
}

func writeMinimalBundle(t *testing.T, root, tracePath string) string {
	t.Helper()
	now := time.Date(2026, 7, 4, 20, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	if err := os.MkdirAll(filepath.Join(root, "projections"), 0o755); err != nil {
		t.Fatalf("mkdir projections: %v", err)
	}
	writeJSON(t, filepath.Join(root, "manifest.json"), map[string]any{
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
	writeJSON(t, filepath.Join(root, "flow.json"), map[string]any{
		"schema":  "acpx.flow-definition-snapshot.v1",
		"name":    "test-flow",
		"startAt": "node-1",
		"nodes": map[string]any{
			"node-1": map[string]any{"nodeType": "compute"},
		},
		"edges": []any{},
	})
	writeJSON(t, filepath.Join(root, "projections", "run.json"), map[string]any{
		"runId":     "run-test",
		"flowName":  "test-flow",
		"startedAt": now,
		"updatedAt": now,
		"status":    "completed",
	})
	writeJSON(t, filepath.Join(root, "projections", "live.json"), map[string]any{
		"runId":     "run-test",
		"flowName":  "test-flow",
		"startedAt": now,
		"updatedAt": now,
		"status":    "completed",
	})
	writeJSON(t, filepath.Join(root, "projections", "steps.json"), []map[string]any{{
		"attemptId":  "node-1#1",
		"nodeId":     "node-1",
		"nodeType":   "compute",
		"outcome":    "ok",
		"startedAt":  now,
		"finishedAt": now,
	}})
	if !strings.HasPrefix(tracePath, "../") {
		trace := strings.Join([]string{
			`{"seq":1,"at":"` + now + `","scope":"node","type":"node_started","runId":"run-test","nodeId":"node-1","attemptId":"node-1#1","payload":{}}`,
			`{"seq":2,"at":"` + now + `","scope":"node","type":"node_outcome","runId":"run-test","nodeId":"node-1","attemptId":"node-1#1","payload":{"outcome":"ok"}}`,
		}, "\n") + "\n"
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(tracePath)), []byte(trace), 0o600); err != nil {
			t.Fatalf("write trace: %v", err)
		}
	}
	return root
}

func writeJSON(t *testing.T, path string, value any) {
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
