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
		assert.Contains(output, "Taskman manages signed-in users and their task lists")
		assert.Contains(output, "auth")
		assert.Contains(output, "completion")
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
		assert.Contains(err.Error(), "taskman auth register <username>")
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
		var whoAmI map[string]any
		err = json.Unmarshal([]byte(output), &whoAmI)
		assert.NoError(err)
		assert.Equal("alice", whoAmI["username"])
	})

	It("creates tasks from prompts when flags are omitted", func() {
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

		output, err := executeCmdWithInput(
			command,
			"Write docs\nDocument the create flow\n",
			"--data-dir", dataDir,
			"create",
		)

		assert.NoError(err)
		assert.Contains(output, "Write docs")
		assert.Contains(output, "Document the create flow")
		assert.Contains(output, "task-1")
		assert.NotContains(output, "incomplete")
		assert.NotContains(output, "\"title\"")
	})

	It("renders created tasks with the task id in the centered footer", func() {
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

		output, err := executeCmd(
			command,
			"--data-dir", dataDir,
			"create",
			"--title", "Write docs",
			"--description", "Document the centered output",
		)

		assert.NoError(err)
		assert.Contains(output, "Write docs")
		assert.Contains(output, "Document the centered output")
		assert.Contains(output, "ID:")
		assert.Contains(output, "task-1")
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
		var updatedTask map[string]any
		err = json.Unmarshal([]byte(output), &updatedTask)
		assert.NoError(err)
		assert.Equal("task-1", updatedTask["id"])
		assert.Equal("complete", updatedTask["completed"])
	})

	It("gets a selected task when no id is provided", func() {
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
			"--description", "Show styled get output",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\n",
			"--data-dir", dataDir,
			"get",
		)
		assert.NoError(err)
		assert.Contains(output, "Write docs")
		assert.Contains(output, "Show styled get output")
		assert.Contains(output, "incomplete")
		assert.NotContains(output, "\"id\"")
		assert.NotContains(output, "\"title\"")
	})

	It("renders retrieved tasks with centered paragraph content and highlighted status", func() {
		dataDir := GinkgoT().TempDir()
		taskStore := tasks.NewStore(dataDir, time.Now, func() string { return "task-1" })
		command := cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore: func(string, func() time.Time, func() string) *tasks.Store {
				return taskStore
			},
			Now:       time.Now,
			NewTaskID: func() string { return "task-1" },
		})

		_, err := executeCmd(command, "--data-dir", dataDir, "auth", "register", "alice")
		assert.NoError(err)

		_, err = executeCmd(command, "--data-dir", dataDir, "auth", "signin", "alice")
		assert.NoError(err)

		_, err = taskStore.Create("alice", tasks.CreateTaskInput{
			Title:       "Write docs",
			Description: "line one\nline two",
		})
		assert.NoError(err)

		output, err := executeCmd(command, "--data-dir", dataDir, "get", "task-1")

		assert.NoError(err)
		assert.Contains(output, "Write docs")
		assert.Contains(output, "line one")
		assert.Contains(output, "line two")
		assert.Contains(output, "STATUS:")
		assert.Contains(output, "incomplete")
		assert.Contains(output, "STATUS:  incomplete")
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

		_, err = executeCmdWithInput(
			command,
			"\n\n\n",
			"--data-dir", dataDir,
			"update", "task-1",
		)

		assert.Error(err)
		assert.Contains(err.Error(), "invalid argument: provide at least one value to update")
		assert.Contains(err.Error(), "taskman update [task-id] --help")
	})

	It("updates a selected task with form values when no id and flags are provided", func() {
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
			"--description", "Cover update form flow",
		)
		assert.NoError(err)

		output, err := executeCmdWithInput(
			command,
			"1\nWrite better tests\nCover form update flow\ntrue\n",
			"--data-dir", dataDir,
			"update",
		)
		assert.NoError(err)
		var updatedTask map[string]any
		err = json.Unmarshal([]byte(output), &updatedTask)
		assert.NoError(err)
		assert.Equal("task-1", updatedTask["id"])
		assert.Equal("Write better tests", updatedTask["title"])
		assert.Equal("Cover form update flow", updatedTask["description"])
		assert.Equal("complete", updatedTask["completed"])
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
		assert.Contains(err.Error(), "taskman list --help")
	})

	It("generates Carapace completion scripts", func() {
		output, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		}), "completion", "powershell")

		assert.NoError(err)
		assert.Contains(output, "_carapace")
		assert.Contains(output, "taskman")
	})

	It("rejects unsupported completion shells", func() {
		_, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{
			NewAuthService: auth.NewService,
			NewTaskStore:   tasks.NewStore,
			Now:            time.Now,
			NewTaskID:      tasks.NewTaskID,
		}), "completion", "cmd")

		assert.Error(err)
		assert.Contains(err.Error(), "shell must be one of [bash elvish fish nushell oil powershell tcsh xonsh zsh]")
		assert.Contains(err.Error(), "taskman completion [shell] --help")
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
		var deletedTask map[string]any
		err = json.Unmarshal([]byte(output), &deletedTask)
		assert.NoError(err)
		assert.Equal("deleted", deletedTask["status"])
		assert.Equal("task-1", deletedTask["id"])
	})
})
