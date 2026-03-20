package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/louiss0/cobra-cli-template/output"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "Manage local task-list users",
		Long:    "Register users, sign in, sign out, and inspect the current session.",
		GroupID: "auth",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
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
		Args:  usernameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]
			err := getAuthServiceFromCommandContext(cmd).Register(username)
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, map[string]string{
				"status":   "registered",
				"username": username,
			})
		},
	}
}

func NewSignInCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "signin <username>",
		Short: "Sign in as a registered user",
		Args:  usernameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]
			err := getAuthServiceFromCommandContext(cmd).SignIn(username)
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, map[string]string{
				"status":   "signed-in",
				"username": username,
			})
		},
	}
}

func NewSignOutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "signout",
		Short: "Sign out the current user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := getAuthServiceFromCommandContext(cmd).SignOut()
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, map[string]string{
				"status": "signed-out",
			})
		},
	}
}

func NewWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the active user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := getAuthServiceFromCommandContext(cmd).CurrentUser()
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, map[string]string{
				"username": username,
			})
		},
	}
}

func usernameArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	if args[0] != lower(args[0]) {
		return fmt.Errorf("argument %q must be lowercase", args[0])
	}

	return nil
}

func writeJSONOutput(cmd *cobra.Command, value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return output.WriteModeAwareOutput(cmd, string(content))
}
