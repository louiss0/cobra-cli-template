package output_test

import (
	"bytes"

	"github.com/charmbracelet/log"
	"github.com/louiss0/cobra-cli-template/output"
	"github.com/louiss0/g-tools/mode"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
)

var _ = Describe("WriteModeAwareOutput", func() {
	var (
		stdout  *bytes.Buffer
		stderr  *bytes.Buffer
		command *cobra.Command
	)

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
		command = &cobra.Command{}
		command.SetOut(stdout)
		command.SetErr(stderr)
	})

	It("writes message and keyvals to stdout in development mode", func() {
		if !mode.NewModeOperator().IsDevelopmentMode() {
			Skip("requires development mode")
		}

		err := output.WriteModeAwareOutput(command, "dev message", "source", "test", "count", 2)

		assert.NoError(err)
		assert.Contains(stdout.String(), "dev message")
		assert.Contains(stdout.String(), "source=test")
		assert.Contains(stdout.String(), "count=2")
		assert.Equal("", stderr.String())
	})

	It("writes json values to stdout in development mode", func() {
		if !mode.NewModeOperator().IsDevelopmentMode() {
			Skip("requires development mode")
		}

		err := output.WriteJSONOutput(command, map[string]string{
			"status": "ok",
		})

		assert.NoError(err)
		assert.Equal("{\"status\":\"ok\"}", stdout.String())
	})

	It("writes message and keyvals to stderr in production mode", func() {
		if !mode.NewModeOperator().IsProductionMode() {
			Skip("requires production mode")
		}

		originalLogger := log.Default()
		testLogger := log.New(stderr)
		log.SetDefault(testLogger)
		defer log.SetDefault(originalLogger)

		err := output.WriteModeAwareOutput(command, "prod message", "source", "test")

		assert.NoError(err)
		assert.Equal("", stdout.String())
		assert.Contains(stderr.String(), "prod message")
		assert.Contains(stderr.String(), "source=test")
	})
})
