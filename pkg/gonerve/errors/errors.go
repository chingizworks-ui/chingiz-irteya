package errors

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func New(msg string) error {
	return errors.New(msg)
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
