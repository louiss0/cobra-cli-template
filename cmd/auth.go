package cmd

import (
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/cobra-cli-template/validation"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "Manage local task-list users",
		Long:    "Register users, sign in, sign out, and inspect the current session.",
		GroupID: "auth",
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		}),
	}

	cmd.AddCommand(
		NewRegisterCmd(),
		NewSignInCmd(),
		NewSignOutCmd(),
		NewWhoAmICmd(),
	)

	return cmd
}

func NewRegisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register <username>",
		Short: "Register a new user",
		Args:  custom_errors.WrapArgs(usernameArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			username := args[0]
			err := getAuthServiceFromCommandContext(cmd).Register(username)
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, map[string]string{
				"status":   "registered",
				"username": username,
			})
		}),
	}
}

func NewSignInCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "signin <username>",
		Short: "Sign in as a registered user",
		Args:  custom_errors.WrapArgs(usernameArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			username := args[0]
			err := getAuthServiceFromCommandContext(cmd).SignIn(username)
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, map[string]string{
				"status":   "signed-in",
				"username": username,
			})
		}),
	}
}

func NewSignOutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "signout",
		Short: "Sign out the current user",
		Args:  custom_errors.WrapArgs(cobra.NoArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			err := getAuthServiceFromCommandContext(cmd).SignOut()
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, map[string]string{
				"status": "signed-out",
			})
		}),
	}
}

func NewWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the active user",
		Args:  custom_errors.WrapArgs(cobra.NoArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			username, err := getAuthServiceFromCommandContext(cmd).CurrentUser()
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, map[string]string{
				"username": username,
			})
		}),
	}
}

func usernameArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	return validation.ValidateLowercaseText("username", args[0])
}
