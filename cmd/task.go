package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/cobra-cli-template/custom_flags"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			title, description, err := runCreateTaskForm(cmd, flags.Title, flags.Description)
			if err != nil {
				return err
			}

			task, err := getTaskStoreFromCommandContext(cmd).Create(username, tasks.CreateTaskInput{
				Title:       title,
				Description: description,
			})
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, tasks.PresentTask(task))
		},
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "Task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "Task description.")

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

			return output.WriteJSONOutput(cmd, tasks.PresentTasks(taskList))
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

			return output.WriteJSONOutput(cmd, tasks.PresentTask(task))
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
		Use:   "update [task-id]",
		Short: "Update an existing task",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
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

			taskID, err := resolveTaskIDForAction(cmd, username, args, "update")
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

			task, err := getTaskStoreFromCommandContext(cmd).Update(username, taskID, updateInput)
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, tasks.PresentTask(task))
		},
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "New task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "New task description.")
	cmd.Flags().Var(&completedFlag, "completed", "Set the completion state.")

	return cmd
}

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [task-id]",
		Short: "Delete a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			taskID, err := resolveTaskIDForAction(cmd, username, args, "delete")
			if err != nil {
				return err
			}

			err = getTaskStoreFromCommandContext(cmd).Delete(username, taskID)
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, map[string]string{
				"status": "deleted",
				"id":     taskID,
			})
		},
	}
}

func runCreateTaskForm(cmd *cobra.Command, currentTitle string, currentDescription string) (string, string, error) {
	if shouldUseInteractiveForm(cmd.InOrStdin()) {
		return runInteractiveCreateTaskForm(cmd, currentTitle, currentDescription)
	}

	return runPromptCreateTaskForm(cmd, currentTitle, currentDescription)
}

func runInteractiveCreateTaskForm(cmd *cobra.Command, currentTitle string, currentDescription string) (string, string, error) {
	title := currentTitle
	description := currentDescription

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Description("Enter the task title.").
				Value(&title).
				Validate(requiredText("title")),
			huh.NewInput().
				Title("Description").
				Description("Enter the task description.").
				Value(&description).
				Validate(requiredText("description")),
		).Title("Create Task"),
	).Run()
	if err != nil {
		return "", "", fmt.Errorf("run task form: %w", err)
	}

	return title, description, nil
}

func runPromptCreateTaskForm(cmd *cobra.Command, currentTitle string, currentDescription string) (string, string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())

	title, err := runPromptInput(cmd, reader, "Title", currentTitle)
	if err != nil {
		return "", "", err
	}

	description, err := runPromptInput(cmd, reader, "Description", currentDescription)
	if err != nil {
		return "", "", err
	}

	return title, description, nil
}

func requiredText(name string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}

		return nil
	}
}

func runPromptInput(cmd *cobra.Command, reader *bufio.Reader, label string, currentValue string) (string, error) {
	if strings.TrimSpace(currentValue) != "" {
		return currentValue, nil
	}

	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s: ", label)
	if err != nil {
		return "", fmt.Errorf("write %s prompt: %w", strings.ToLower(label), err)
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}

	value = strings.TrimSpace(value)
	if err := requiredText(strings.ToLower(label))(value); err != nil {
		return "", err
	}

	return value, nil
}

func resolveTaskIDForAction(cmd *cobra.Command, username string, args []string, action string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	taskList, err := getTaskStoreFromCommandContext(cmd).List(username, tasks.ListFilterAll)
	if err != nil {
		return "", err
	}

	if len(taskList) == 0 {
		return "", fmt.Errorf("no tasks available to %s", action)
	}

	if shouldUseInteractiveForm(cmd.InOrStdin()) {
		return selectTaskIDInteractive(taskList, action)
	}

	return selectTaskIDFromPrompt(cmd, taskList, action)
}

func selectTaskIDInteractive(taskList []tasks.Task, action string) (string, error) {
	selectedTaskID := taskList[0].ID
	options := lo.Map(taskList, func(task tasks.Task, _ int) huh.Option[string] {
		return huh.NewOption(taskSelectionLabel(task), task.ID)
	})

	err := huh.NewSelect[string]().
		Title(actionTitle(action) + " Task").
		Description("Choose a task by title and ID.").
		Options(options...).
		Value(&selectedTaskID).
		Run()
	if err != nil {
		return "", fmt.Errorf("select task to %s: %w", action, err)
	}

	return selectedTaskID, nil
}

func selectTaskIDFromPrompt(cmd *cobra.Command, taskList []tasks.Task, action string) (string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())

	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Select a task to %s:\n", action)
	if err != nil {
		return "", fmt.Errorf("write %s prompt: %w", action, err)
	}

	for index, task := range taskList {
		_, err = fmt.Fprintf(cmd.ErrOrStderr(), "%d. %s\n", index+1, taskSelectionLabel(task))
		if err != nil {
			return "", fmt.Errorf("write task list: %w", err)
		}
	}

	_, err = fmt.Fprint(cmd.ErrOrStderr(), "Choice: ")
	if err != nil {
		return "", fmt.Errorf("write choice prompt: %w", err)
	}

	choiceValue, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read choice: %w", err)
	}

	choiceNumber, err := strconv.Atoi(strings.TrimSpace(choiceValue))
	if err != nil || choiceNumber < 1 || choiceNumber > len(taskList) {
		return "", fmt.Errorf("choice must be between 1 and %d", len(taskList))
	}

	return taskList[choiceNumber-1].ID, nil
}

func shouldUseInteractiveForm(reader io.Reader) bool {
	file, ok := reader.(interface{ Fd() uintptr })
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
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

func taskSelectionLabel(task tasks.Task) string {
	return fmt.Sprintf("%s [%s]", task.Title, task.ID)
}

func actionTitle(action string) string {
	if action == "" {
		return ""
	}

	return strings.ToUpper(action[:1]) + action[1:]
}
