package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

const (
	defaultTaskListWidth  = 80
	defaultTaskListHeight = 20
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

type taskListModel struct {
	list           list.Model
	selectedTaskID string
}

func (model taskListModel) Init() tea.Cmd {
	return nil
}

func (model taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.WindowSizeMsg:
		model.list.SetSize(message.Width, message.Height)
		return model, nil
	case tea.KeyMsg:
		switch message.String() {
		case "enter":
			if selectedTask, ok := model.list.SelectedItem().(taskListItem); ok {
				model.selectedTaskID = selectedTask.task.ID
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

func runTaskListUI(cmd *cobra.Command, taskList []tasks.Task) error {
	_, err := runTaskListSelection(cmd, "Tasks", buildTaskItems(taskList))
	if err != nil {
		return err
	}

	return nil
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
	taskListComponent := list.New(
		items,
		list.NewDefaultDelegate(),
		defaultTaskListWidth,
		defaultTaskListHeight,
	)
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

func buildTaskItems(taskList []tasks.Task) []list.Item {
	items := make([]list.Item, 0, len(taskList))

	for _, task := range taskList {
		items = append(items, taskListItem{task: task})
	}

	return items
}
