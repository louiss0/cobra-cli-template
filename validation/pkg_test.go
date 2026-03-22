package validation_test

import (
	"errors"

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

var _ = Describe("string validation schemas", func() {
	It("rejects empty required text", func() {
		err := validation.ValidateRequiredText("title", "   ")

		assert.Error(err)
		assert.Contains(err.Error(), "title cannot be empty")
	})

	It("rejects non-lowercase text", func() {
		err := validation.ValidateLowercaseText("username", "Alice")

		assert.Error(err)
		assert.Contains(err.Error(), "username must be lowercase")
	})

	It("rejects values outside an allowed set", func() {
		err := validation.ValidateAllowedString("pending", []string{"all", "complete"}, func(message string) error {
			return errors.New(message)
		})

		assert.Error(err)
		assert.Contains(err.Error(), "value must be one of [all complete]")
	})

	It("parses boolean strings", func() {
		value, err := validation.ParseBoolString("true", func(message string) error {
			return errors.New(message)
		})

		assert.NoError(err)
		assert.True(value)
	})

	It("rejects integers outside a range", func() {
		_, err := validation.ParseIntegerRange("11", 1, 10, func(message string) error {
			return errors.New(message)
		})

		assert.Error(err)
		assert.Contains(err.Error(), "value must be between 1 and 10")
	})
})
