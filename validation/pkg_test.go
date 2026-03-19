package validation_test

import (
	"github.com/louiss0/cobra-cli-template/validation"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
)

var _ = Describe("function values struct schema", func() {
	It("accepts structs with non-nil function fields only", func() {
		type dependencies struct {
			CommandRunner func() error
			ContextSetup  func(*cobra.Command, []string) error
		}

		schema := validation.NewFunctionValuesStructSchema[dependencies]()

		_, err := schema.Parse(dependencies{
			CommandRunner: func() error { return nil },
			ContextSetup:  func(*cobra.Command, []string) error { return nil },
		})

		assert.NoError(err)
	})

	It("rejects structs with missing function values", func() {
		type dependencies struct {
			CommandRunner func() error
			ContextSetup  func(*cobra.Command, []string) error
		}

		schema := validation.NewFunctionValuesStructSchema[dependencies]()

		_, err := schema.Parse(dependencies{
			CommandRunner: func() error { return nil },
		})

		assert.Error(err)
		assert.Contains(err.Error(), "missing functions")
		assert.Contains(err.Error(), "ContextSetup")
	})

	It("rejects structs that contain non-function fields", func() {
		type dependencies struct {
			CommandRunner func() error
			Name          string
		}

		schema := validation.NewFunctionValuesStructSchema[dependencies]()

		_, err := schema.Parse(dependencies{
			CommandRunner: func() error { return nil },
			Name:          "template",
		})

		assert.Error(err)
		assert.Contains(err.Error(), "non-function fields")
		assert.Contains(err.Error(), "Name")
	})
})
