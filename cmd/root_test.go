package cmd_test

import (
	"github.com/louiss0/cobra-cli-template/cmd"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Root Command", func() {

	It("runs with args", func() {

		output, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{}), "info")

		assert.NoError(err)
		assert.Equal("Root here info", output)
	})

	It("runs successfully", func() {
		output, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{}))

		assert.NoError(err)
		assert.Equal("Root here", output)
	})

	It("shows help information", func() {
		output, err := executeCmd(cmd.NewRootCmd(cmd.Dependencies{}), "--help")

		assert.NoError(err)
		assert.Contains(output, "starter template for building maintainable Cobra applications")
	})
})
