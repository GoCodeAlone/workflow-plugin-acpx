// Package acpxruntime provides the small ACPX durable-artifact surface shared
// by the Workflow plugin and Go consumers such as ratchet-cli.
package acpxruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	acpx "github.com/GoCodeAlone/acpx-go"
	"github.com/GoCodeAlone/acpx-go/flowjson"
)

// BundleSummary is a content-safe summary of an ACPX durable flow run bundle.
type BundleSummary struct {
	RunID        string `json:"run_id"`
	FlowName     string `json:"flow_name,omitempty"`
	Status       string `json:"status"`
	TraceCount   int    `json:"trace_count"`
	StepCount    int    `json:"step_count"`
	SessionCount int    `json:"session_count"`
}

// FlowSummary is a content-safe summary of an ACPX JSON flow definition.
type FlowSummary struct {
	Name      string   `json:"name,omitempty"`
	StartAt   string   `json:"start_at"`
	NodeCount int      `json:"node_count"`
	Requires  []string `json:"requires,omitempty"`
}

// ValidateBundle verifies that root is an ACPX durable flow run bundle.
func ValidateBundle(ctx context.Context, root string) error {
	_, err := acpx.LoadBundle(ctx, root)
	return err
}

// ReplaySummary validates root and returns a summary safe for logs and step outputs.
func ReplaySummary(ctx context.Context, root string) (BundleSummary, error) {
	bundle, err := acpx.LoadBundle(ctx, root)
	if err != nil {
		return BundleSummary{}, err
	}
	projection, err := acpx.RebuildTraceProjection(bundle.Trace)
	if err != nil {
		return BundleSummary{}, err
	}
	return BundleSummary{
		RunID:        bundle.Manifest.RunID,
		FlowName:     bundle.Manifest.FlowName,
		Status:       string(bundle.Manifest.Status),
		TraceCount:   len(bundle.Trace),
		StepCount:    len(projection.Attempts),
		SessionCount: len(bundle.Manifest.Sessions),
	}, nil
}

// ValidateFlowFile reads and validates an ACPX-compatible JSON flow definition.
func ValidateFlowFile(path string) (FlowSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FlowSummary{}, err
	}
	return ValidateFlow(data)
}

// ValidateFlow validates a serialized ACPX-compatible JSON flow definition.
func ValidateFlow(data []byte) (FlowSummary, error) {
	var def flowjson.Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return FlowSummary{}, fmt.Errorf("%w: %v", flowjson.ErrInvalidDefinition, err)
	}
	if err := def.Validate(); err != nil {
		return FlowSummary{}, err
	}
	return FlowSummary{
		Name:      def.Name,
		StartAt:   def.StartAt,
		NodeCount: len(def.Nodes),
		Requires:  append([]string(nil), def.Requires...),
	}, nil
}
