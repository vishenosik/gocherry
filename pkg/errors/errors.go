package errors

import "github.com/pkg/errors"

func New(msg string) error {
	return errors.New(msg)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...any) error {
	return errors.Wrapf(err, format, args...)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
