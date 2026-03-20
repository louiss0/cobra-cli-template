package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_flags"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

func NewTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "Manage tasks for the signed-in user",
		Long:    "Create, read, update, delete, and filter tasks for the active user.",
		GroupID: "task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewCreateCmd(),
		NewListCmd(),
		NewGetCmd(),
		NewUpdateCmd(),
		NewDeleteCmd(),
	)

	return cmd
}

func NewCreateCmd() *cobra.Command {
	flags := struct {
		Title       string
		Description string
	}{}
	completedFlag := custom_flags.NewBoolFlag("completed")

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(cmd.InOrStdin())

			title, err := taskFieldValue(cmd, reader, "Title", flags.Title)
			if err != nil {
				return err
			}

			description, err := taskFieldValue(cmd, reader, "Description", flags.Description)
			if err != nil {
				return err
			}

			task, err := getTaskStoreFromCommandContext(cmd).Create(username, tasks.CreateTaskInput{
				Title:       title,
				Description: description,
				Completed:   completedFlag.Value(),
			})
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, tasks.NewPublicTask(task))
		},
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "Task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "Task description.")
	cmd.Flags().Var(&completedFlag, "completed", "Create the task as complete.")

	return cmd
}

func NewListCmd() *cobra.Command {
	flags := struct {
		Status string
	}{
		Status: string(tasks.ListFilterAll),
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			filter, err := parseListFilter(flags.Status)
			if err != nil {
				return err
			}

			taskList, err := getTaskStoreFromCommandContext(cmd).List(username, filter)
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, tasks.NewPublicTasks(taskList))
		},
	}

	cmd.Flags().StringVar(&flags.Status, "status", string(tasks.ListFilterAll), "Filter by all, complete, or incomplete.")

	return cmd
}

func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <task-id>",
		Short: "Show one task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			task, err := getTaskStoreFromCommandContext(cmd).Get(username, args[0])
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, tasks.NewPublicTask(task))
		},
	}
}

func NewUpdateCmd() *cobra.Command {
	flags := struct {
		Title       string
		Description string
	}{}
	completedFlag := custom_flags.NewBoolFlag("completed")

	cmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Update an existing task",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if !cmd.Flags().Changed("title") && !cmd.Flags().Changed("description") && !cmd.Flags().Changed("completed") {
				return fmt.Errorf("provide at least one flag to update")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			updateInput := tasks.UpdateTaskInput{}

			if cmd.Flags().Changed("title") {
				updateInput.Title = &flags.Title
			}

			if cmd.Flags().Changed("description") {
				updateInput.Description = &flags.Description
			}

			if cmd.Flags().Changed("completed") {
				updateInput.Completed = tasks.OptionalBool(completedFlag.Value())
			}

			task, err := getTaskStoreFromCommandContext(cmd).Update(username, args[0], updateInput)
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, tasks.NewPublicTask(task))
		},
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "New task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "New task description.")
	cmd.Flags().Var(&completedFlag, "completed", "Set the completion state.")

	return cmd
}

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			err = getTaskStoreFromCommandContext(cmd).Delete(username, args[0])
			if err != nil {
				return err
			}

			return writeJSONOutput(cmd, map[string]string{
				"status": "deleted",
				"id":     args[0],
			})
		},
	}
}

func taskFieldValue(cmd *cobra.Command, reader *bufio.Reader, label string, currentValue string) (string, error) {
	if strings.TrimSpace(currentValue) != "" {
		return currentValue, nil
	}

	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s: ", label)
	if err != nil {
		return "", fmt.Errorf("write prompt: %w", err)
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", strings.ToLower(label))
	}

	return value, nil
}

func parseListFilter(value string) (tasks.ListFilter, error) {
	switch value {
	case string(tasks.ListFilterAll):
		return tasks.ListFilterAll, nil
	case string(tasks.ListFilterComplete):
		return tasks.ListFilterComplete, nil
	case string(tasks.ListFilterIncomplete):
		return tasks.ListFilterIncomplete, nil
	default:
		return "", fmt.Errorf("status must be one of all, complete, incomplete")
	}
}

func lower(value string) string {
	return strings.ToLower(value)
}
