package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/louiss0/cobra-cli-template/auth"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

var supportedCompletionShells = []string{
	"bash",
	"elvish",
	"fish",
	"nushell",
	"oil",
	"powershell",
	"tcsh",
	"xonsh",
	"zsh",
}

type registeredUsersDocument struct {
	Users []string `json:"users"`
}

func NewCompletionCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate Carapace shell completions",
		Long: "Generate Taskman shell completions powered by Carapace. " +
			"When no shell is provided, Carapace will try to detect it automatically.",
		Example: `taskman completion bash
taskman completion powershell | Out-String | Invoke-Expression
taskman _carapace zsh | source`,
		Args: custom_errors.WrapArgs(validateCompletionArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			shell := ""
			if len(args) == 1 {
				shell = args[0]
			}

			snippet, err := carapace.Gen(root).Snippet(shell)
			if err != nil {
				if shell == "" {
					return fmt.Errorf("generate completion snippet: %w", err)
				}

				return fmt.Errorf("generate %s completion snippet: %w", shell, err)
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), snippet)
			if err != nil {
				return fmt.Errorf("write completion snippet: %w", err)
			}

			return nil
		}),
	}

	carapace.Gen(cmd).PositionalCompletion(actionSupportedShells())

	return cmd
}

func configureRootAutocomplete(cmd *cobra.Command) {
	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"data-dir": carapace.ActionDirectories(),
	})
}

func validateCompletionArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return err
	}

	if len(args) == 0 {
		return nil
	}

	if slices.Contains(supportedCompletionShells, args[0]) {
		return nil
	}

	return custom_errors.CreateInvalidArgumentErrorWithMessage(
		fmt.Sprintf(
			"shell must be one of [%s]",
			strings.Join(supportedCompletionShells, " "),
		),
	)
}

func actionSupportedShells() carapace.Action {
	return carapace.ActionValuesDescribed(
		"bash", "Bash",
		"elvish", "Elvish",
		"fish", "Fish",
		"nushell", "Nushell",
		"oil", "Oil shell",
		"powershell", "PowerShell",
		"tcsh", "Tcsh",
		"xonsh", "Xonsh",
		"zsh", "Zsh",
	)
}

func actionTaskStatusValues() carapace.Action {
	return carapace.ActionValuesDescribed(
		tasks.ListFilterAll, "Show complete and incomplete tasks",
		tasks.ListFilterComplete, "Show only completed tasks",
		tasks.ListFilterIncomplete, "Show only incomplete tasks",
	)
}

func actionCompletionStateValues() carapace.Action {
	return carapace.ActionValuesDescribed(
		"true", "Mark the task as complete",
		"false", "Mark the task as incomplete",
	)
}

func actionRegisteredUsers(command *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		users, err := loadRegisteredUsers(dataDirFromCommand(command))
		if err != nil || len(users) == 0 {
			return carapace.ActionValues()
		}

		return carapace.ActionValues(users...)
	})
}

func actionTaskIDs(command *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		dataDir := dataDirFromCommand(command)
		username, err := auth.NewService(dataDir).CurrentUser()
		if err != nil {
			return carapace.ActionValues()
		}

		taskList, err := tasks.NewStore(dataDir, time.Now, tasks.NewTaskID).List(username, tasks.ListFilterAll)
		if err != nil || len(taskList) == 0 {
			return carapace.ActionValues()
		}

		values := make([]string, 0, len(taskList)*2)
		for _, task := range taskList {
			values = append(values, task.ID, task.Title)
		}

		return carapace.ActionValuesDescribed(values...)
	})
}

func dataDirFromCommand(command *cobra.Command) string {
	flag := command.Root().Flag("data-dir")
	if flag == nil {
		return defaultDataDir()
	}

	value := flag.Value.String()
	if value == "" {
		return defaultDataDir()
	}

	return value
}

func loadRegisteredUsers(dataDir string) ([]string, error) {
	content, err := os.ReadFile(filepath.Join(dataDir, "users.json"))
	if err != nil {
		return nil, err
	}

	var document registeredUsersDocument
	if err := json.Unmarshal(content, &document); err != nil {
		return nil, err
	}

	return document.Users, nil
}
