package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/louiss0/cobra-cli-template/build_info"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/cobra-cli-template/validation"
	"github.com/spf13/cobra"
)

type commandContextKey string

func GenerateContextFromMap(cmd *cobra.Command, dependencies map[commandContextKey]any) context.Context {

	ctx := cmd.Context()
	for k, v := range dependencies {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

type Dependencies struct {
	CommandRunner func() error
	ContextSetup  func(*cobra.Command, []string) error
}

var rootCmd *cobra.Command

var schema = validation.NewFunctionValuesStructSchema[Dependencies]()

func init() {
	rootCmd = NewRootCmd(Dependencies{})
}

func NewRootCmd(deps Dependencies) *cobra.Command {

	err := validateDependencies(deps)
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:   "cli",
		Short: "Build CLI applications with Cobra and Go",
		Long: `A starter template for building maintainable Cobra applications.
The template is organized for test-driven development using Ginkgo and
Testify assertions.`,
		Version: build_info.Version(),

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			ctx := GenerateContextFromMap(cmd, map[commandContextKey]any{})

			cmd.SetContext(ctx)

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			message := "Root here"
			if len(args) > 0 {
				message = message + " " + strings.Join(args, " ")
			}

			return output.WriteModeAwareOutput(cmd, message)
		},
	}

	return cmd
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func validateDependencies(deps Dependencies) error {
	if deps.CommandRunner == nil && deps.ContextSetup == nil {
		return nil
	}

	_, err := schema.Parse(deps)
	if err != nil {
		return fmt.Errorf("invalid root dependencies: %w", err)
	}

	return nil
}
