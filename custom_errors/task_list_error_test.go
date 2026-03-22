package custom_errors_test

import (
	"fmt"

	"github.com/louiss0/cobra-cli-template/auth"
	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
)

var _ = Describe("task list errors", func() {
	It("formats not signed in errors with sign-in guidance", func() {
		err := custom_errors.CreateTaskListError(auth.ErrNotSignedIn)

		assert.Contains(err.Error(), "taskman error: sign-in required")
		assert.Contains(err.Error(), "taskman auth signin <username>")
	})

	It("formats unknown user errors with register guidance", func() {
		err := custom_errors.CreateTaskListError(fmt.Errorf("%w: alice", auth.ErrUserNotFound))

		assert.Contains(err.Error(), "taskman error: user not found: alice")
		assert.Contains(err.Error(), "taskman auth register <username>")
	})

	It("formats task not found errors with list guidance", func() {
		err := custom_errors.CreateTaskListError(fmt.Errorf("%w: task-404", tasks.ErrTaskNotFound))

		assert.Contains(err.Error(), "taskman error: task not found: task-404")
		assert.Contains(err.Error(), "taskman list")
	})

	It("wraps run errors using the task list formatter", func() {
		runE := custom_errors.WrapRunE(func(cmd *cobra.Command, args []string) error {
			return auth.ErrNotSignedIn
		})

		err := runE(&cobra.Command{}, []string{})

		assert.Error(err)
		assert.Contains(err.Error(), "taskman error: sign-in required")
	})

	It("wraps args validation errors using the task list formatter", func() {
		validate := custom_errors.WrapArgs(cobra.ExactArgs(1))
		rootCommand := &cobra.Command{Use: "taskman"}
		authCommand := &cobra.Command{Use: "auth"}
		registerCommand := &cobra.Command{Use: "register <username>"}
		rootCommand.AddCommand(authCommand)
		authCommand.AddCommand(registerCommand)

		err := validate(registerCommand, []string{})

		assert.Error(err)
		assert.Contains(err.Error(), "taskman error: accepts 1 arg(s), received 0")
		assert.Contains(err.Error(), "provide required argument(s): <username>")
		assert.Contains(err.Error(), "taskman auth register <username>")
	})
})
