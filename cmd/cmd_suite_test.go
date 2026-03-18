package cmd_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	tAssert "github.com/stretchr/testify/assert"
)

var assert *tAssert.Assertions

func TestCobraCliTemplate(t *testing.T) {
	assert = tAssert.New(GinkgoT())
	RunSpecs(t, "CMD Suite")
}

func executeCmd(command *cobra.Command, args ...string) (string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	command.SetOut(stdout)
	command.SetErr(stderr)
	command.SetArgs(args)

	err := command.Execute()
	if err != nil && stderr.Len() > 0 {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), err
}
