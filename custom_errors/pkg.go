// Package custom_errors provides reusable, explicit errors for CLI validation.
package custom_errors

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidFlag     = errors.New("invalid flag")
	ErrInvalidArgument = errors.New("invalid argument")
)

var validFlagNamePattern = regexp.MustCompile(`^[a-z0-9]+$`)

type FlagName string

func (name FlagName) Validate() error {
	if !validFlagNamePattern.MatchString(string(name)) {
		return fmt.Errorf(
			"%w %q: name must be lowercase alphanumeric",
			ErrInvalidFlag,
			name,
		)
	}

	return nil
}

func CreateInvalidFlagErrorWithMessage(flagName FlagName, message string) error {
	if err := flagName.Validate(); err != nil {
		return err
	}

	return fmt.Errorf("%w %q: %s", ErrInvalidFlag, flagName, message)
}

func CreateInvalidArgumentErrorWithMessage(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidArgument, message)
}
