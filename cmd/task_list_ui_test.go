package cmd

import (
	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
	tAssert "github.com/stretchr/testify/assert"
)

var _ = Describe("Task List UI", func() {
	It("shows add task option for all and incomplete filters", func() {
		assert := tAssert.New(GinkgoT())

		assert.True(shouldShowAddTaskOption(tasks.ListFilterAll))
		assert.True(shouldShowAddTaskOption(tasks.ListFilterIncomplete))
	})

	It("does not show add task option for complete filter", func() {
		assert := tAssert.New(GinkgoT())

		assert.False(shouldShowAddTaskOption(tasks.ListFilterComplete))
	})

	It("adds an add-task item when the filter supports creating tasks", func() {
		assert := tAssert.New(GinkgoT())

		taskList := []tasks.Task{
			{ID: "task-1", Title: "Write tests", Description: "Cover list UI", Completed: false},
		}

		items := buildTaskListItems(taskList, tasks.ListFilterAll)

		assert.Len(items, 2)
		_, ok := items[1].(addTaskListItem)
		assert.True(ok)
	})

	It("does not add the add-task item for complete filter", func() {
		assert := tAssert.New(GinkgoT())

		taskList := []tasks.Task{
			{ID: "task-1", Title: "Write tests", Description: "Cover list UI", Completed: true},
		}

		items := buildTaskListItems(taskList, tasks.ListFilterComplete)

		assert.Len(items, 1)
		_, ok := items[0].(taskListItem)
		assert.True(ok)
	})
})
