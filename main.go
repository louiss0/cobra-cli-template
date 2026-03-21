package main

import (
	"os"

	"github.com/louiss0/cobra-cli-template/cmd"
	"github.com/louiss0/cobra-cli-template/output"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_ = output.WriteModeAwareError(os.Stderr, err)
		os.Exit(1)
	}
}
