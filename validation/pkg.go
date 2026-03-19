package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod"
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
