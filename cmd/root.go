package cmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/louiss0/cobra-cli-template/auth"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/louiss0/cobra-cli-template/validation"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type commandContextKey string

const (
	_AUTH_SERVICE commandContextKey = "auth-service"
	_TASK_STORE   commandContextKey = "task-store"
	_CURRENT_TIME commandContextKey = "current-time"
)

func GenerateContextFromMap(cmd *cobra.Command, dependencies map[string]any) context.Context {
	return lo.Reduce(
		lo.Entries(dependencies),
		func(ctx context.Context, entry lo.Entry[string, any], _ int) context.Context {
			return context.WithValue(ctx, entry.Key, entry.Value)
		},
		cmd.Context(),
	)
}

type Dependencies struct {
	NewAuthService func(string) *auth.Service
	NewTaskStore   func(string, func() time.Time, func() string) *tasks.Store
	Now            func() time.Time
	NewTaskID      func() string
}

var rootCmd *cobra.Command

var schema = validation.NewFunctionValuesStructSchema[Dependencies]()

func init() {
	rootCmd = NewRootCmd(Dependencies{
		NewAuthService: auth.NewService,
		NewTaskStore:   tasks.NewStore,
		Now:            time.Now,
		NewTaskID:      tasks.NewTaskID,
	})
}

func NewRootCmd(deps Dependencies) *cobra.Command {
	if _, err := schema.Parse(deps); err != nil {
		panic(err)
	}

	rootFlags := struct {
		DataDir string
	}{}

	cmd := &cobra.Command{
		Use:           "task-list",
		Short:         "Manage signed-in users and their task lists",
		Long:          "Manage signed-in users and their task lists with local JSON storage.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			authService := deps.NewAuthService(rootFlags.DataDir)
			taskStore := deps.NewTaskStore(rootFlags.DataDir, deps.Now, deps.NewTaskID)

			ctx := GenerateContextFromMap(cmd, map[string]any{
				string(_AUTH_SERVICE): authService,
				string(_TASK_STORE):   taskStore,
				string(_CURRENT_TIME): deps.Now,
			})

			cmd.SetContext(ctx)
			return nil
		},
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		}),
	}

	cmd.PersistentFlags().StringVar(
		&rootFlags.DataDir,
		"data-dir",
		defaultDataDir(),
		"Directory used for users, session, and task JSON files.",
	)

	cmd.AddGroup(
		&cobra.Group{ID: "auth", Title: "Authentication Commands"},
		&cobra.Group{ID: "task", Title: "Task Commands"},
	)

	cmd.AddCommand(
		NewAuthCmd(),
		NewCreateCmd(),
		NewListCmd(),
		NewGetCmd(),
		NewUpdateCmd(),
		NewDeleteCmd(),
	)

	return cmd
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func defaultDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", ".task-list")
	}

	return filepath.Join(configDir, "task-list")
}

func getAuthServiceFromCommandContext(cmd *cobra.Command) *auth.Service {
	return cmd.Context().Value(string(_AUTH_SERVICE)).(*auth.Service)
}

func getTaskStoreFromCommandContext(cmd *cobra.Command) *tasks.Store {
	return cmd.Context().Value(string(_TASK_STORE)).(*tasks.Store)
}

func signedInUsernameFromCommandContext(cmd *cobra.Command) (string, error) {
	username, err := getAuthServiceFromCommandContext(cmd).CurrentUser()
	if err != nil {
		return "", err
	}

	return username, nil
}
