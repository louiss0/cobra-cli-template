package cmd

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
	tAssert "github.com/stretchr/testify/assert"
)

var _ = Describe("Task List UI", func() {
	It("builds task items without an add-task entry", func() {
		assert := tAssert.New(GinkgoT())

		items := buildTaskItems([]tasks.Task{
			{ID: "task-1", Title: "Write tests", Description: "Cover list UI", Completed: false},
			{ID: "task-2", Title: "Ship CLI", Description: "Review task list", Completed: true},
		})

		assert.Len(items, 2)

		for _, item := range items {
			_, ok := item.(taskListItem)
			assert.True(ok)
		}
	})

	It("renders task items after the list receives a window size", func() {
		assert := tAssert.New(GinkgoT())

		model := taskListModel{
			list: listWithItems(
				buildTaskItems([]tasks.Task{
					{ID: "task-1", Title: "Write tests", Description: "Cover list UI"},
				}),
				0,
				0,
			),
		}

		nextModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

		view := nextModel.(taskListModel).View()

		assert.True(strings.Contains(view, "Write tests"))
		assert.True(strings.Contains(view, "Cover list UI"))
	})
})

func listWithItems(items []list.Item, width int, height int) list.Model {
	taskList := list.New(items, list.NewDefaultDelegate(), width, height)
	taskList.Title = "Tasks"
	taskList.SetShowHelp(true)
	taskList.SetFilteringEnabled(false)
	taskList.SetStatusBarItemName("task", "tasks")
	return taskList
}
