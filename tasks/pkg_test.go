package tasks_test

import (
	"encoding/json"
	"time"

	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Store", func() {
	It("creates and filters tasks by completion state", func() {
		dataDir := GinkgoT().TempDir()
		now := time.Date(2026, time.March, 20, 9, 0, 0, 0, time.UTC)
		store := tasks.NewStore(dataDir, func() time.Time { return now }, func() string { return "task-1" })

		createdTask, err := store.Create("alice", tasks.CreateTaskInput{
			Title:       "Write tests",
			Description: "Cover the create flow",
		})
		assert.NoError(err)

		updatedTask, err := store.Update("alice", createdTask.ID, tasks.UpdateTaskInput{
			Title:       createdTask.Title,
			Description: createdTask.Description,
			Completed:   true,
		})
		assert.NoError(err)
		assert.True(updatedTask.Completed)
		assert.Equal(now.Unix(), updatedTask.CreatedAt())
		assert.Equal(now.Unix(), updatedTask.UpdatedAt())

		allTasks, err := store.List("alice", tasks.ListFilterAll)
		assert.NoError(err)
		assert.Len(allTasks, 1)

		completeTasks, err := store.List("alice", tasks.ListFilterComplete)
		assert.NoError(err)
		assert.Len(completeTasks, 1)

		incompleteTasks, err := store.List("alice", tasks.ListFilterIncomplete)
		assert.NoError(err)
		assert.Empty(incompleteTasks)
	})

	It("removes completed tasks during weekly cleanup", func() {
		dataDir := GinkgoT().TempDir()
		now := time.Date(2026, time.March, 20, 9, 0, 0, 0, time.UTC)
		currentTime := now
		store := tasks.NewStore(dataDir, func() time.Time { return currentTime }, func() string { return "task-1" })

		createdTask, err := store.Create("alice", tasks.CreateTaskInput{
			Title:       "Ship feature",
			Description: "Before the deadline",
			Completed:   true,
		})
		assert.NoError(err)
		assert.True(createdTask.Completed)

		currentTime = now.Add(8 * 24 * time.Hour)

		listedTasks, err := store.List("alice", tasks.ListFilterAll)

		assert.NoError(err)
		assert.Empty(listedTasks)
	})

	It("presents tasks with date strings through an any value", func() {
		dataDir := GinkgoT().TempDir()
		now := time.Date(2026, time.March, 20, 9, 0, 0, 0, time.UTC)
		store := tasks.NewStore(dataDir, func() time.Time { return now }, func() string { return "task-1" })

		task, err := store.Create("alice", tasks.CreateTaskInput{
			Title:       "Write docs",
			Description: "Explain the output",
		})
		assert.NoError(err)

		content, err := json.Marshal(tasks.PresentTask(task))
		assert.NoError(err)

		var presentedTask map[string]any
		err = json.Unmarshal(content, &presentedTask)

		assert.NoError(err)
		assert.Equal("incomplete", presentedTask["completed"])
		assert.Equal(now.Format(time.RFC3339), presentedTask["createdAt"])
		assert.Equal(now.Format(time.RFC3339), presentedTask["updatedAt"])
	})
})
