# workflow-plugin-acpx

Workflow plugin for ACPX durable bundle validation and summaries.

This plugin exposes Go-native ACPX replay/archive checks as Workflow steps. It
wraps `github.com/GoCodeAlone/acpx-go`; it does not execute `.flow.ts` files
and it does not implement an ACP transport.

## Installation

```sh
wfctl plugin install workflow-plugin-acpx
```

## Step Types

- `acpx.bundle_validate` validates an ACPX durable flow run bundle.
- `acpx.bundle_summary` validates a bundle and returns safe summary counts.
- `acpx.flow_validate` validates an ACPX-compatible JSON flow definition.

## Development

```sh
make build
make test
wfctl plugin validate-contract --for-publish --tag v0.1.0 .
```

## Shared Go Runtime

Go consumers can import the small helper package without the Workflow SDK:

```go
summary, err := acpxruntime.ReplaySummary(ctx, runDir)
```

