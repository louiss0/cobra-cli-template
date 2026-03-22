package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/coerce"
	"github.com/kaptinlin/gozod/types"
	"github.com/louiss0/cobra-cli-template/custom_errors"
)

func NewFunctionValuesStructSchema[T any]() *gozod.ZodStruct[T, T] {
	return gozod.FromStruct[T]().Check(func(value T, payload *gozod.ParsePayload) {
		invalidFields, missingFunctions := collectFunctionFieldViolations(value)

		if len(invalidFields) > 0 {
			payload.AddIssueWithMessage(
				fmt.Sprintf(
					"gozod validation failed: non-function fields: %s",
					strings.Join(invalidFields, ", "),
				),
			)
		}

		if len(missingFunctions) > 0 {
			payload.AddIssueWithMessage(
				fmt.Sprintf(
					"gozod validation failed: missing functions: %s",
					strings.Join(missingFunctions, ", "),
				),
			)
		}
	})
}

func collectFunctionFieldViolations(input any) ([]string, []string) {
	value := reflect.ValueOf(input)

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return []string{"<root>"}, []string{}
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return []string{"<root>"}, []string{}
	}

	typ := value.Type()
	invalidFields := make([]string, 0, typ.NumField())
	missingFunctions := make([]string, 0, typ.NumField())

	for fieldIndex := range typ.NumField() {
		fieldType := typ.Field(fieldIndex)
		fieldValue := value.Field(fieldIndex)

		if fieldType.Type.Kind() != reflect.Func {
			invalidFields = append(invalidFields, fieldType.Name)
			continue
		}

		if fieldValue.IsNil() {
			missingFunctions = append(missingFunctions, fieldType.Name)
		}
	}

	return invalidFields, missingFunctions
}

func NewRequiredTextSchema(name string) *gozod.ZodString[string] {
	return gozod.String().
		Trim().
		Min(1, fmt.Sprintf("%s cannot be empty", name))
}

func ValidateRequiredText(name string, value string) error {
	_, err := NewRequiredTextSchema(name).Parse(value)
	return err
}

func NewLowercaseTextSchema(name string) *gozod.ZodString[string] {
	return gozod.String().Lowercase(fmt.Sprintf("%s must be lowercase", name))
}

func ValidateLowercaseText(name string, value string) error {
	_, err := NewLowercaseTextSchema(name).Parse(value)
	if err != nil {
		return custom_errors.CreateInvalidArgumentErrorWithMessage(
			fmt.Sprintf("%s must be lowercase", name),
		)
	}

	return nil
}

func NewAllowedStringSchema(allowedValues []string) *gozod.ZodEnum[string, string] {
	return gozod.EnumSlice(allowedValues)
}

func ValidateAllowedString(value string, allowedValues []string, errorFactory func(string) error) error {
	_, err := NewAllowedStringSchema(allowedValues).Parse(value)
	if err != nil {
		return errorFactory(fmt.Sprintf("value must be one of %v", allowedValues))
	}

	return nil
}

func NewBoolStringSchema() *gozod.ZodStringBool[bool] {
	return coerce.StringBool("value must be either true or false")
}

func ParseBoolString(value string, errorFactory func(string) error) (bool, error) {
	parsedValue, err := NewBoolStringSchema().Parse(value)
	if err != nil {
		return false, errorFactory("value must be either true or false")
	}

	return parsedValue, nil
}

func NewIntegerRangeSchema(min int, max int) *types.ZodIntegerTyped[int, int] {
	return coerce.Int(
		fmt.Sprintf("value must be an integer between %d and %d", min, max),
	).Min(
		int64(min),
		fmt.Sprintf("value must be between %d and %d", min, max),
	).Max(
		int64(max),
		fmt.Sprintf("value must be between %d and %d", min, max),
	)
}

func ParseIntegerRange(value string, min int, max int, errorFactory func(string) error) (int, error) {
	number, err := NewIntegerRangeSchema(min, max).Parse(value)
	if err != nil {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue == "" {
			return 0, errorFactory(
				fmt.Sprintf("value must be an integer between %d and %d", min, max),
			)
		}

		parsedNumber, parseErr := coerce.Int().Parse(trimmedValue)
		if parseErr != nil {
			return 0, errorFactory(
				fmt.Sprintf("value must be an integer between %d and %d", min, max),
			)
		}

		if parsedNumber < min || parsedNumber > max {
			return 0, errorFactory(
				fmt.Sprintf("value must be between %d and %d", min, max),
			)
		}
	}

	return number, nil
}
