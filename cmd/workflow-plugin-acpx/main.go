package main

import (
	"fmt"
	"os"

	"github.com/GoCodeAlone/workflow-plugin-acpx/internal"
	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("workflow-plugin-acpx: Workflow external plugin for ACPX durable bundle validation and summaries.")
		return
	}
	sdk.Serve(internal.NewProvider(),
		sdk.WithBuildVersion(sdk.ResolveBuildVersion(internal.Version)),
	)
}
