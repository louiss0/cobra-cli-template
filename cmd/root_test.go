package cmd_test

import (
	"encoding/json"
	"time"

	"github.com/louiss0/cobra-cli-template/auth"
	"github.com/louiss0/cobra-cli-template/cmd"
	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Root Command", func() {
	It("shows help information", func() {
		output, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		}), "--help")

		assert.NoError(err)
		assert.Contains(output, "Manage signed-in users and their task lists")
		assert.Contains(output, "auth")
		assert.Contains(output, "task")
	})

	It("requires sign-in before managing tasks", func() {
		dataDir := GinkgoT().TempDir()

		_, err := executeCmd(
			cmd.NewRootCmd(cmd.Dependencies{
				NewAuthService: auth.NewService,
				NewTaskStore:   tasks.NewStore,
				Now:            time.Now,
				NewTaskID:      tasks.NewTaskID,
			}),
			"--data-dir", dataDir,
			"task", "list",
		)

		assert.Error(err)
		assert.Contains(err.Error(), "sign in first")
	})

	It("registers, signs in, and reports the active user", func() {
		dataDir := GinkgoT().TempDir()
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		output, err := executeCmd(command, "--data-dir", dataDir, "auth", "whoami")

		assert.NoError(err)
		assert.Contains(output, "\"username\":\"alice\"")
	})

	It("creates tasks from prompts when flags are omitted", func() {
		dataDir := GinkgoT().TempDir()
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"Write docs\nDocument the create flow\n",
			"--data-dir", dataDir,
			"task", "create",
		)

		assert.NoError(err)
		assert.Contains(output, "\"title\":\"Write docs\"")
		assert.Contains(output, "\"description\":\"Document the create flow\"")
		assert.Contains(output, "\"completed\":\"incomplete\"")
	})

	It("lists complete and incomplete tasks separately", func() {
		dataDir := GinkgoT().TempDir()
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      func() string { return "task-1" },
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		_, err = executeCmd(
			command,
			"--data-dir", dataDir,
			"task", "create",
			"--title", "Write tests",
			"--description", "Cover the list flow",
		)
		assert.NoError(err)

		_, err = executeCmd(
			command,
			"--data-dir", dataDir,
			"task", "update", "task-1",
			"--completed", "true",
		)
		assert.NoError(err)

		output, err := executeCmd(
			command,
			"--data-dir", dataDir,
			"task", "list",
			"--status", "complete",
		)
		assert.NoError(err)

		var listedTasks []map[string]any
		err = json.Unmarshal([]byte(output), &listedTasks)

		assert.NoError(err)
		assert.Len(listedTasks, 1)
		assert.Equal("complete", listedTasks[0]["completed"])
	})

	It("updates a selected task when no id is provided", func() {
		dataDir := GinkgoT().TempDir()
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      func() string { return "task-1" },
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		_, err = executeCmd(
			command,
			"--data-dir", dataDir,
			"task", "create",
			"--title", "Write tests",
			"--description", "Cover selection flow",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\n",
			"--data-dir", dataDir,
			"task", "update",
			"--completed", "true",
		)
		assert.NoError(err)
		assert.Contains(output, "\"id\":\"task-1\"")
		assert.Contains(output, "\"completed\":\"complete\"")
	})

	It("deletes a selected task when no id is provided", func() {
		dataDir := GinkgoT().TempDir()
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      func() string { return "task-1" },
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		_, err = executeCmd(
			command,
			"--data-dir", dataDir,
			"task", "create",
			"--title", "Write docs",
			"--description", "Cover delete selection flow",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\n",
			"--data-dir", dataDir,
			"task", "delete",
		)
		assert.NoError(err)
		assert.Contains(output, "\"status\":\"deleted\"")
		assert.Contains(output, "\"id\":\"task-1\"")
	})
})
