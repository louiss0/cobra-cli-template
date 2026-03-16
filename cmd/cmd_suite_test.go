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

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {

	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	err := cmd.Execute()

	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}
