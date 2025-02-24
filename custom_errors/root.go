package custom_errors

import (
	"errors"
	"fmt"
	"regexp"
)

var InvalidFlag = errors.New("Invalid Flag:")

var InvalidArgument = errors.New("Invalid Argument:")

type FlagName string

func (self FlagName) Error() error {

	regex := regexp.MustCompile(`^[a-z0-9]+$`)

	if !regex.MatchString(string(self)) {
		return fmt.Errorf("%w %s a flag name must be alphanumeric from start to end %s", InvalidFlag, self, string(self))
	}

	return nil
}

var CreateInvalidFlagErrorWithMessage = func(flagName FlagName, message string) error {

	if err := flagName.Error(); err != nil {
		return err
	}

	return fmt.Errorf("%w %s %s", InvalidFlag, flagName, message)

}

var CreateInvalidArgumentErrorWithMessage = func(message string) error {

	return fmt.Errorf("%w %s", InvalidFlag, message)

}
