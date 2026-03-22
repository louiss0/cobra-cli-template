package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/custom_flags"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/louiss0/g-tools/mode"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var modeOperator = mode.NewModeOperator()

func NewCreateCmd() *cobra.Command {
	flags := struct {
		Title       string
		Description string
	}{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a task",
		GroupID: "task",
		Args:    custom_errors.WrapArgs(cobra.NoArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
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

			return writeCreatedTaskOutput(cmd, task)
		}),
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "Task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "Task description.")

	return cmd
}

func NewListCmd() *cobra.Command {
	flags := struct {
		Status string
	}{
		Status: tasks.ListFilterAll,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		GroupID: "task",
		Args:    custom_errors.WrapArgs(cobra.NoArgs),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
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

			if shouldUseTaskListUI(cmd) {
				return runTaskListUI(cmd, username, filter, taskList)
			}

			return output.WriteJSONOutput(cmd, tasks.PresentTasks(taskList))
		}),
	}

	cmd.Flags().StringVar(&flags.Status, "status", tasks.ListFilterAll, "Filter by all, complete, or incomplete.")

	return cmd
}

func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get [task-id]",
		Short:   "Show one task",
		GroupID: "task",
		Args:    custom_errors.WrapArgs(cobra.MaximumNArgs(1)),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			taskID, err := resolveTaskIDForGet(cmd, username, args)
			if err != nil {
				return err
			}

			task, err := getTaskStoreFromCommandContext(cmd).Get(username, taskID)
			if err != nil {
				return err
			}

			return writeStyledTaskOutput(cmd, task)
		}),
	}
}

func NewUpdateCmd() *cobra.Command {
	flags := struct {
		Title       string
		Description string
	}{}
	completedFlag := custom_flags.NewBoolFlag("completed")

	cmd := &cobra.Command{
		Use:     "update [task-id]",
		Short:   "Update an existing task",
		GroupID: "task",
		Args: custom_errors.WrapArgs(func(cmd *cobra.Command, args []string) error {
			return cobra.MaximumNArgs(1)(cmd, args)
		}),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			var promptReader *bufio.Reader
			if !shouldUseInteractiveForm(cmd.InOrStdin()) {
				promptReader = bufio.NewReader(cmd.InOrStdin())
			}

			username, err := signedInUsernameFromCommandContext(cmd)
			if err != nil {
				return err
			}

			taskID, err := resolveTaskIDForActionWithReader(cmd, username, args, "update", promptReader)
			if err != nil {
				return err
			}

			updateInput, err := resolveUpdateInput(
				cmd,
				username,
				taskID,
				flags.Title,
				flags.Description,
				completedFlag.Value(),
				promptReader,
			)
			if err != nil {
				return err
			}

			task, err := getTaskStoreFromCommandContext(cmd).Update(username, taskID, updateInput)
			if err != nil {
				return err
			}

			return output.WriteJSONOutput(cmd, tasks.PresentTask(task))
		}),
	}

	cmd.Flags().StringVar(&flags.Title, "title", "", "New task title.")
	cmd.Flags().StringVar(&flags.Description, "description", "", "New task description.")
	cmd.Flags().Var(&completedFlag, "completed", "Set the completion state.")

	return cmd
}

func resolveUpdateInput(
	cmd *cobra.Command,
	username string,
	taskID string,
	title string,
	description string,
	completed bool,
	promptReader *bufio.Reader,
) (tasks.UpdateTaskInput, error) {
	task, err := getTaskStoreFromCommandContext(cmd).Get(username, taskID)
	if err != nil {
		return tasks.UpdateTaskInput{}, err
	}

	nextTitle := task.Title
	nextDescription := task.Description
	nextCompleted := task.Completed

	hasChanges := false
	if cmd.Flags().Changed("title") {
		nextTitle = title
		hasChanges = true
	}

	if cmd.Flags().Changed("description") {
		nextDescription = description
		hasChanges = true
	}

	if cmd.Flags().Changed("completed") {
		nextCompleted = completed
		hasChanges = true
	}

	if hasChanges {
		return createUpdateInputFromTaskDiff(task, nextTitle, nextDescription, nextCompleted)
	}

	return runUpdateTaskForm(cmd, task, promptReader)
}

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [task-id]",
		Short:   "Delete a task",
		GroupID: "task",
		Args:    custom_errors.WrapArgs(cobra.MaximumNArgs(1)),
		RunE: custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
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
		}),
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
			huh.NewText().
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

func runUpdateTaskForm(cmd *cobra.Command, task tasks.Task, promptReader *bufio.Reader) (tasks.UpdateTaskInput, error) {
	if shouldUseInteractiveForm(cmd.InOrStdin()) {
		return runInteractiveUpdateTaskForm(task)
	}

	return runPromptUpdateTaskForm(cmd, task, promptReader)
}

func runInteractiveUpdateTaskForm(task tasks.Task) (tasks.UpdateTaskInput, error) {
	title := task.Title
	description := task.Description
	completed := task.Completed

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Description("Update the task title.").
				Value(&title).
				Validate(requiredText("title")),
			huh.NewText().
				Title("Description").
				Description("Update the task description.").
				Value(&description).
				Validate(requiredText("description")),
			huh.NewConfirm().
				Title("Completed").
				Description("Mark the task as completed.").
				Value(&completed),
		).Title("Update Task"),
	).Run()
	if err != nil {
		return tasks.UpdateTaskInput{}, fmt.Errorf("run task update form: %w", err)
	}

	return createUpdateInputFromTaskDiff(task, title, description, completed)
}

func runPromptUpdateTaskForm(cmd *cobra.Command, task tasks.Task, reader *bufio.Reader) (tasks.UpdateTaskInput, error) {
	title, err := runPromptOptionalInput(cmd, reader, "Title", task.Title)
	if err != nil {
		return tasks.UpdateTaskInput{}, err
	}

	description, err := runPromptOptionalInput(cmd, reader, "Description", task.Description)
	if err != nil {
		return tasks.UpdateTaskInput{}, err
	}

	completed, err := runPromptOptionalBoolInput(cmd, reader, "Completed", task.Completed)
	if err != nil {
		return tasks.UpdateTaskInput{}, err
	}

	return createUpdateInputFromTaskDiff(task, title, description, completed)
}

func createUpdateInputFromTaskDiff(
	task tasks.Task,
	title string,
	description string,
	completed bool,
) (tasks.UpdateTaskInput, error) {
	updateInput := tasks.UpdateTaskInput{
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
	}
	changedCount := 0

	if title != task.Title {
		updateInput.Title = title
		changedCount++
	}

	if description != task.Description {
		updateInput.Description = description
		changedCount++
	}

	if completed != task.Completed {
		updateInput.Completed = completed
		changedCount++
	}

	if changedCount == 0 {
		return tasks.UpdateTaskInput{}, custom_errors.CreateInvalidArgumentErrorWithMessage(
			"provide at least one value to update",
		)
	}

	return updateInput, nil
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

func runPromptOptionalInput(cmd *cobra.Command, reader *bufio.Reader, label string, currentValue string) (string, error) {
	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s [%s]: ", label, currentValue)
	if err != nil {
		return "", fmt.Errorf("write %s prompt: %w", strings.ToLower(label), err)
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return currentValue, nil
	}

	if err := requiredText(strings.ToLower(label))(value); err != nil {
		return "", err
	}

	return value, nil
}

func runPromptOptionalBoolInput(cmd *cobra.Command, reader *bufio.Reader, label string, currentValue bool) (bool, error) {
	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s (true/false) [%t]: ", label, currentValue)
	if err != nil {
		return false, fmt.Errorf("write %s prompt: %w", strings.ToLower(label), err)
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return currentValue, nil
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, custom_errors.CreateInvalidArgumentErrorWithMessage(
			fmt.Sprintf("%s must be true or false", strings.ToLower(label)),
		)
	}

	return parsedValue, nil
}

func resolveTaskIDForAction(cmd *cobra.Command, username string, args []string, action string) (string, error) {
	return resolveTaskIDForActionWithReader(cmd, username, args, action, nil)
}

func resolveTaskIDForGet(cmd *cobra.Command, username string, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	taskList, err := getTaskStoreFromCommandContext(cmd).List(username, tasks.ListFilterAll)
	if err != nil {
		return "", err
	}

	if len(taskList) == 0 {
		return "", fmt.Errorf("no tasks available to get")
	}

	if shouldUseTaskListUI(cmd) {
		return runTaskIDSelectionUI(cmd, taskList, "Select Task")
	}

	return selectTaskIDFromPrompt(cmd, bufio.NewReader(cmd.InOrStdin()), taskList, "get")
}

func resolveTaskIDForActionWithReader(
	cmd *cobra.Command,
	username string,
	args []string,
	action string,
	reader *bufio.Reader,
) (string, error) {
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

	if reader == nil {
		reader = bufio.NewReader(cmd.InOrStdin())
	}

	return selectTaskIDFromPrompt(cmd, reader, taskList, action)
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

func selectTaskIDFromPrompt(cmd *cobra.Command, reader *bufio.Reader, taskList []tasks.Task, action string) (string, error) {
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
	return isTerminal(reader)
}

func shouldUseTaskListUI(cmd *cobra.Command) bool {
	if !modeOperator.IsProductionMode() {
		return false
	}

	return isTerminal(cmd.InOrStdin()) && isTerminal(cmd.OutOrStdout())
}

func isTerminal(value any) bool {
	file, ok := value.(interface{ Fd() uintptr })
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
}

func parseListFilter(value string) (string, error) {
	switch value {
	case tasks.ListFilterAll:
		return tasks.ListFilterAll, nil
	case tasks.ListFilterComplete:
		return tasks.ListFilterComplete, nil
	case tasks.ListFilterIncomplete:
		return tasks.ListFilterIncomplete, nil
	default:
		return "", custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName("status"),
			"value must be one of [all complete incomplete]",
		)
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
