package custom_flags_test

import (
	"github.com/louiss0/cobra-cli-template/custom_flags"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("custom flags", func() {
	It("rejects empty string values", func() {
		flag := custom_flags.NewEmptyStringFlag("name")

		err := flag.Set("   ")

		assert.Error(err)
	})

	It("parses bool values", func() {
		flag := custom_flags.NewBoolFlag("enabled")

		err := flag.Set("true")

		assert.NoError(err)
		assert.True(flag.Value())
	})

	It("enforces allowed values for union flag", func() {
		flag := custom_flags.NewUnionFlag([]string{"dev", "prod"}, "mode")

		err := flag.Set("qa")

		assert.Error(err)
	})

	It("enforces numeric ranges", func() {
		flag := custom_flags.NewRangeFlag("count", 1, 10)

		err := flag.Set("11")

		assert.Error(err)
	})
})
