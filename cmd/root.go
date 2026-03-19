package cmd

import (
	"fmt"

	"github.com/louiss0/cobra-cli-template/output"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {

	var Args struct {
		Info string
	}

	cmd := &cobra.Command{
		Use:   "cli",
		Short: "Build CLI applications with Cobra and Go",
		Long: `A starter template for building maintainable Cobra applications.
The template is organized for test-driven development using Ginkgo and
Testify assertions.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				Args.Info = args[0]
				return output.WriteModeAwareOutput(cmd, fmt.Sprintf("Root here %s", Args.Info))

			}

			return output.WriteModeAwareOutput(cmd, "Root here")
		},
	}

	cmd.MarkFlagsMutuallyExclusive()

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
