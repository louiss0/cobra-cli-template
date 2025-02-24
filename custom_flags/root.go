package customFlags

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
