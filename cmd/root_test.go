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
		assert.Contains(output, "create")
		assert.Contains(output, "list")
	})

	It("does not expose task as a nested sub-command", func() {
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		})

		_, err := executeCmd(command, "task", "list")

		assert.Error(err)
		assert.Contains(err.Error(), "unknown command \"task\"")
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
			"list",
		)

		assert.Error(err)
		assert.Contains(err.Error(), "sign-in required")
	})

	It("shows the required username argument for auth register", func() {
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		})

		_, err := executeCmd(command, "auth", "register")

		assert.Error(err)
		assert.Contains(err.Error(), "accepts 1 arg(s), received 0")
		assert.Contains(err.Error(), "provide required argument(s): <username>")
		assert.Contains(err.Error(), "task-list auth register <username>")
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
			"create",
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
			"create",
			"--title", "Write tests",
			"--description", "Cover the list flow",
		)
		assert.NoError(err)

		_, err = executeCmd(
			command,
			"--data-dir", dataDir,
			"update", "task-1",
			"--completed", "true",
		)
		assert.NoError(err)

		output, err := executeCmd(
			command,
			"--data-dir", dataDir,
			"list",
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
			"create",
			"--title", "Write tests",
			"--description", "Cover selection flow",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\n",
			"--data-dir", dataDir,
			"update",
			"--completed", "true",
		)
		assert.NoError(err)
		assert.Contains(output, "\"id\":\"task-1\"")
		assert.Contains(output, "\"completed\":\"complete\"")
	})

	It("shows a clear flag message when update has no changes", func() {
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
			"create",
			"--title", "Write tests",
			"--description", "Cover update validation",
		)
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "update", "task-1")

		assert.Error(err)
		assert.Contains(err.Error(), "invalid argument: provide at least one flag to update")
		assert.Contains(err.Error(), "task-list update [task-id] --help")
	})

	It("shows allowed status values for list filtering", func() {
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

		_, err = executeCmd(command, "--data-dir", dataDir, "list", "--status", "pending")

		assert.Error(err)
		assert.Contains(err.Error(), "invalid flag \"status\": value must be one of [all complete incomplete]")
		assert.Contains(err.Error(), "task-list list --help")
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
			"create",
			"--title", "Write docs",
			"--description", "Cover delete selection flow",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\n",
			"--data-dir", dataDir,
			"delete",
		)
		assert.NoError(err)
		assert.Contains(output, "\"status\":\"deleted\"")
		assert.Contains(output, "\"id\":\"task-1\"")
	})
})
