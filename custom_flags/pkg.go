// Package custom_flags provides custom pflag.Value implementations.
package custom_flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/louiss0/cobra-cli-template/custom_errors"
)

type emptyStringFlag struct {
	value    string
	flagName string
}

func NewEmptyStringFlag(flagName string) emptyStringFlag {
	return emptyStringFlag{flagName: flagName}
}

func (flag emptyStringFlag) String() string {
	return flag.value
}

func (flag *emptyStringFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName(flag.flagName),
			"value cannot be empty",
		)
	}

	flag.value = value
	return nil
}

func (flag emptyStringFlag) Type() string {
	return "string"
}

type boolFlag struct {
	value    bool
	flagName string
}

func NewBoolFlag(flagName string) boolFlag {
	return boolFlag{flagName: flagName}
}

func (flag boolFlag) String() string {
	return strconv.FormatBool(flag.value)
}

func (flag *boolFlag) Set(value string) error {
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName(flag.flagName),
			"value must be either true or false",
		)
	}

	flag.value = parsedValue
	return nil
}

func (flag boolFlag) Type() string {
	return "bool"
}

func (flag boolFlag) Value() bool {
	return flag.value
}

type unionFlag struct {
	value         string
	allowedValues []string
	flagName      string
}

func NewUnionFlag(allowedValues []string, flagName string) unionFlag {
	return unionFlag{
		allowedValues: allowedValues,
		flagName:      flagName,
	}
}

func (flag unionFlag) String() string {
	return flag.value
}

func (flag *unionFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName(flag.flagName),
			"value cannot be empty",
		)
	}

	for _, allowedValue := range flag.allowedValues {
		if value == allowedValue {
			flag.value = value
			return nil
		}
	}

	return custom_errors.CreateInvalidFlagErrorWithMessage(
		custom_errors.FlagName(flag.flagName),
		fmt.Sprintf("value must be one of %v", flag.allowedValues),
	)
}

func (flag unionFlag) Type() string {
	return "string"
}

type RangeFlag struct {
	value    int
	min      int
	max      int
	flagName string
}

func NewRangeFlag(flagName string, min, max int) RangeFlag {
	if min > max {
		panic("min must be less than or equal to max")
	}

	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}

	return RangeFlag{
		min:      min,
		max:      max,
		flagName: flagName,
	}
}

func (flag RangeFlag) String() string {
	return strconv.Itoa(flag.value)
}

func (flag RangeFlag) Value() int {
	return flag.value
}

func (flag *RangeFlag) Set(value string) error {
	number, err := strconv.Atoi(value)
	if err != nil {
		return custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName(flag.flagName),
			fmt.Sprintf("value must be an integer between %d and %d", flag.min, flag.max),
		)
	}

	if number < flag.min || number > flag.max {
		return custom_errors.CreateInvalidFlagErrorWithMessage(
			custom_errors.FlagName(flag.flagName),
			fmt.Sprintf("value must be between %d and %d", flag.min, flag.max),
		)
	}

	flag.value = number
	return nil
}

func (flag RangeFlag) Type() string {
	return "int"
}
