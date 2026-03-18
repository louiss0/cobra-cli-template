package custom_errors_test

import (
	"errors"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("custom errors", func() {
	It("validates flag names", func() {
		err := custom_errors.FlagName("bad-flag").Validate()

		assert.Error(err)
		assert.True(errors.Is(err, custom_errors.ErrInvalidFlag))
	})

	It("returns argument errors with the correct base error", func() {
		err := custom_errors.CreateInvalidArgumentErrorWithMessage("missing id")

		assert.Error(err)
		assert.True(errors.Is(err, custom_errors.ErrInvalidArgument))
	})
})
