package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

type taskListItem struct {
	task tasks.Task
}

func (item taskListItem) FilterValue() string {
	return fmt.Sprintf("%s %s", item.task.Title, item.task.Description)
}

func (item taskListItem) Title() string {
	status := "incomplete"
	if item.task.Completed {
		status = "complete"
	}

	return fmt.Sprintf("%s [%s]", item.task.Title, status)
}

func (item taskListItem) Description() string {
	return fmt.Sprintf("%s (id: %s)", item.task.Description, item.task.ID)
}

type addTaskListItem struct{}

func (item addTaskListItem) FilterValue() string {
	return "add task create new task"
}

func (item addTaskListItem) Title() string {
	return "Add task"
}

func (item addTaskListItem) Description() string {
	return "Create a new task from this list"
}

type taskListModel struct {
	list           list.Model
	shouldAddTask  bool
	selectedTaskID string
}

func (model taskListModel) Init() tea.Cmd {
	return nil
}

func (model taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "enter":
			if selectedTask, ok := model.list.SelectedItem().(taskListItem); ok {
				model.selectedTaskID = selectedTask.task.ID
			}

			if _, ok := model.list.SelectedItem().(addTaskListItem); ok {
				model.shouldAddTask = true
			}

			return model, tea.Quit
		case "esc", "q", "ctrl+c":
			return model, tea.Quit
		}
	}

	var command tea.Cmd
	model.list, command = model.list.Update(msg)
	return model, command
}

func (model taskListModel) View() string {
	return model.list.View()
}

func runTaskListUI(cmd *cobra.Command, username string, filter string, taskList []tasks.Task) error {
	selection, err := runTaskListSelection(cmd, "Tasks", buildTaskListItems(taskList, filter))
	if err != nil {
		return err
	}

	if !selection.shouldAddTask {
		return nil
	}

	title, description, err := runCreateTaskForm(cmd, "", "")
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
}

func runTaskIDSelectionUI(cmd *cobra.Command, taskList []tasks.Task, title string) (string, error) {
	selection, err := runTaskListSelection(cmd, title, buildTaskItems(taskList))
	if err != nil {
		return "", err
	}

	if selection.selectedTaskID == "" {
		return "", fmt.Errorf("no task selected")
	}

	return selection.selectedTaskID, nil
}

func runTaskListSelection(cmd *cobra.Command, title string, items []list.Item) (taskListModel, error) {
	taskListComponent := list.New(items, list.NewDefaultDelegate(), 0, 0)
	taskListComponent.Title = title
	taskListComponent.SetShowHelp(true)
	taskListComponent.SetFilteringEnabled(false)
	taskListComponent.SetStatusBarItemName("task", "tasks")

	model := taskListModel{
		list: taskListComponent,
	}

	program := tea.NewProgram(
		model,
		tea.WithInput(cmd.InOrStdin()),
		tea.WithOutput(cmd.OutOrStdout()),
	)

	finalModel, err := program.Run()
	if err != nil {
		return taskListModel{}, fmt.Errorf("run task list: %w", err)
	}

	return finalModel.(taskListModel), nil
}

func buildTaskListItems(taskList []tasks.Task, filter string) []list.Item {
	items := buildTaskItems(taskList)

	if shouldShowAddTaskOption(filter) {
		items = append(items, addTaskListItem{})
	}

	return items
}

func buildTaskItems(taskList []tasks.Task) []list.Item {
	items := make([]list.Item, 0, len(taskList)+1)

	for _, task := range taskList {
		items = append(items, taskListItem{task: task})
	}

	return items
}

func shouldShowAddTaskOption(filter string) bool {
	return filter == tasks.ListFilterAll || filter == tasks.ListFilterIncomplete
}
