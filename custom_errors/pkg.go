// Package custom_errors provides reusable, explicit errors for CLI validation.
package custom_errors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/louiss0/cobra-cli-template/auth"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

var (
	ErrInvalidFlag     = errors.New("invalid flag")
	ErrInvalidArgument = errors.New("invalid argument")
)

const (
	signInCommand   = "taskman auth signin <username>"
	registerCommand = "taskman auth register <username>"
)

var validFlagNamePattern = regexp.MustCompile(`^[a-z0-9]+$`)

type FlagName string

func (name FlagName) Validate() error {
	_, err := gozod.String().
		Regex(validFlagNamePattern, "name must be lowercase alphanumeric").
		Parse(string(name))
	if err != nil {
		return fmt.Errorf("%w %q: name must be lowercase alphanumeric", ErrInvalidFlag, name)
	}

	return nil
}

func CreateInvalidFlagErrorWithMessage(flagName FlagName, message string) error {
	if err := flagName.Validate(); err != nil {
		return err
	}

	return fmt.Errorf("%w %q: %s", ErrInvalidFlag, flagName, message)
}

func CreateInvalidArgumentErrorWithMessage(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidArgument, message)
}

func WrapRunE(runE func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(command *cobra.Command, args []string) error {
		err := runE(command, args)
		if err != nil {
			return CreateTaskListCommandError(command, err)
		}

		return nil
	}
}

func WrapArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		err := validate(command, args)
		if err != nil {
			return CreateTaskListCommandError(command, err)
		}

		return nil
	}
}

func CreateTaskListError(err error) error {
	return CreateTaskListCommandError(nil, err)
}

func CreateTaskListCommandError(command *cobra.Command, err error) error {
	switch {
	case errors.Is(err, auth.ErrNotSignedIn):
		return createTaskListErrorWithHint(
			"sign-in required",
			fmt.Sprintf("run `%s`", signInCommand),
		)
	case errors.Is(err, auth.ErrUserNotFound):
		return createTaskListErrorWithHint(
			err.Error(),
			fmt.Sprintf("user not found; register first with `%s`", registerCommand),
		)
	case errors.Is(err, auth.ErrUserExists):
		return createTaskListErrorWithHint(
			err.Error(),
			fmt.Sprintf("user already exists; sign in with `%s`", signInCommand),
		)
	case errors.Is(err, tasks.ErrTaskNotFound):
		return createTaskListErrorWithHint(
			err.Error(),
			"run `taskman list` to find a valid task id",
		)
	case isCobraArgumentError(err):
		return createTaskListErrorWithHint(
			err.Error(),
			createArgumentHint(command),
		)
	case errors.Is(err, ErrInvalidFlag):
		return createTaskListErrorWithHint(
			err.Error(),
			createCommandHelpHint(command, "run `taskman <command> --help` for valid flag values"),
		)
	case errors.Is(err, ErrInvalidArgument):
		return createTaskListErrorWithHint(
			err.Error(),
			createCommandHelpHint(command, "run `taskman <command> --help` for argument usage"),
		)
	default:
		return createTaskListErrorWithHint(
			err.Error(),
			createCommandHelpHint(command, ""),
		)
	}
}

func createTaskListErrorWithHint(message string, hint string) error {
	if hint == "" {
		return fmt.Errorf("taskman error: %s", message)
	}

	return fmt.Errorf("taskman error: %s\nhint: %s", message, hint)
}

func isCobraArgumentError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "arg(s)") && strings.Contains(message, "received")
}

func createArgumentHint(command *cobra.Command) string {
	invocation := createCommandInvocation(command)
	requiredArguments := extractRequiredArguments(command)

	if len(requiredArguments) > 0 {
		return fmt.Sprintf(
			"provide required argument(s): %s; example: `%s`",
			strings.Join(requiredArguments, ", "),
			invocation,
		)
	}

	return fmt.Sprintf("check usage: `%s`", invocation)
}

func createCommandHelpHint(command *cobra.Command, fallback string) string {
	invocation := createCommandInvocation(command)
	if invocation == "taskman" {
		return fallback
	}

	return fmt.Sprintf("run `%s --help` for usage details", invocation)
}

func createCommandInvocation(command *cobra.Command) string {
	if command == nil {
		return "taskman"
	}

	useTokens := strings.Fields(command.Use)
	if len(useTokens) <= 1 {
		return command.CommandPath()
	}

	return fmt.Sprintf("%s %s", command.CommandPath(), strings.Join(useTokens[1:], " "))
}

func extractRequiredArguments(command *cobra.Command) []string {
	if command == nil {
		return nil
	}

	useTokens := strings.Fields(command.Use)
	if len(useTokens) <= 1 {
		return nil
	}

	return filterByPredicate(
		useTokens[1:],
		func(token string) bool {
			return strings.HasPrefix(token, "<") && strings.HasSuffix(token, ">")
		},
	)
}

func filterByPredicate(values []string, predicate func(string) bool) []string {
	filteredValues := []string{}

	for _, value := range values {
		if predicate(value) {
			filteredValues = append(filteredValues, value)
		}
	}

	return filteredValues
}
