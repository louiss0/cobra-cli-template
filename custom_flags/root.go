package custom_flags

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/samber/lo"
)

type emptyStringFlag struct {
	value    string
	flagName string
}

func NewEmptyStringFlag(flagName string) emptyStringFlag {
	return emptyStringFlag{
		flagName: flagName,
	}
}

func (t emptyStringFlag) String() string {
	return t.value
}

func (t *emptyStringFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\s+$`, value)

	if error != nil {
		return error
	}

	if match {
		return fmt.Errorf(
			"The %s is empty",
			t.flagName,
		)
	}

	t.value = value
	return nil
}

func (t emptyStringFlag) Type() string {
	return "string"
}

type boolFlag struct {
	value    string
	flagName string
}

func NewBoolFlag(flagName string) boolFlag {
	return boolFlag{
		flagName: flagName,
	}
}

func (c boolFlag) String() string {
	return c.value
}

func (c *boolFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains([]string{"true", "false"}, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(c.flagName),
			[]string{"true", "false"},
		)
	}
	c.value = value
	return nil
}

func (c boolFlag) Type() string {
	return "bool"
}

func (c boolFlag) Value() bool {
	value, _ := strconv.ParseBool(c.value)
	return value

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

func (self unionFlag) String() string {
	return self.value
}

func (self *unionFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains(self.allowedValues, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(self.flagName),
			self.allowedValues,
		)

	}
	self.value = value
	return nil
}

func (self unionFlag) Type() string {
	return "string"
}

type RangeFlag struct {
	value, min, max int
	flagName        string
}

func NewRangeFlag(flagName string, min, max int) RangeFlag {

	if min > max {
		panic("min must be less than max")
	}

	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}

	if min > max {
		panic("min must be less than max")
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

func (self RangeFlag) String() string {

	return fmt.Sprintf("%d", self.value)
}

func (self RangeFlag) Value() int {
	return self.value
}

func (self *RangeFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\d+$`, value)

	if error != nil {
		return error
	}

	if match {
		num, _ := strconv.Atoi(value)
		if num < self.min || num > self.max {
			return fmt.Errorf(
				"%sflag must be between %d and %d",
				custom_errors.FlagName(self.flagName),
				self.min,
				self.max,
			)
		}
		self.value = num
		return nil
	}

	return fmt.Errorf(
		"%sflag must be an integer between %d and %d",
		custom_errors.FlagName(self.flagName),
		self.min,
		self.max,
	)
}

func (self RangeFlag) Type() string {
	return "string"
}
