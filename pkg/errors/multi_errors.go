package errors

import (
	"encoding/json"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type MultiError struct {
	critical error
	errs     *multierror.Error
}

func (er *MultiError) ErrorOrNil() error {
	if er.errs == nil || len(er.errs.Errors) == 0 {
		return nil
	}
	return er
}

func (er *MultiError) Error() string {
	return er.errs.Error()
}

func (er *MultiError) Unwrap() error {
	return er.errs.Unwrap()
}

func (er *MultiError) Critical() error {
	return er.critical
}

func (er *MultiError) CriticalString() string {
	if er.critical != nil {
		return er.critical.Error()
	}
	return ""
}

func (er *MultiError) List() []string {
	if er.errs == nil {
		return nil
	}

	var errors []string

	for _, err := range er.errs.Errors {
		errors = append(errors, err.Error())
	}
	return errors
}

func (er *MultiError) append(err error, critical bool, wrapper func(error) error) {
	if err == nil {
		return
	}

	if critical {
		er.critical = err
	}

	switch err := err.(type) {
	case *multierror.Error:
		for _, _err := range err.Errors {
			er.errs = multierror.Append(er.errs, wrapper(_err))
		}
	case *MultiError:
		er.critical = err.critical
		er.errs = multierror.Append(er.errs, wrapper(err.errs))
	default:
		er.errs = multierror.Append(er.errs, wrapper(err))
	}
}

func (er *MultiError) Append(err error) {
	er.append(err, false, func(err error) error { return err })
}

func (er *MultiError) AppendWrap(err error, message string) {
	er.append(err, false, func(err error) error { return errors.Wrap(err, message) })
}

func (er *MultiError) AppendWrapf(err error, format string, args ...any) {
	er.append(err, false, func(err error) error { return errors.Wrapf(err, format, args...) })
}

func (er *MultiError) AppendCritical(err error) {
	er.append(err, true, func(err error) error { return err })
}

func (er *MultiError) AppendCriticalWrap(err error, message string) {
	er.append(err, true, func(err error) error { return errors.Wrap(err, message) })
}

func (er *MultiError) AppendCriticalWrapf(err error, format string, args ...any) {
	er.append(err, true, func(err error) error { return errors.Wrapf(err, format, args...) })
}

type marshalableMultiError struct {
	Critical string   `yaml:"critical,omitempty" json:"critical,omitempty"`
	Errors   []string `yaml:"errors,omitempty" json:"errors,omitempty"`
}

func toMarshalable(er *MultiError) marshalableMultiError {
	return marshalableMultiError{
		Critical: er.CriticalString(),
		Errors:   er.List(),
	}
}

func (er *MultiError) MarshalJSON() ([]byte, error) {
	if er.errs == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(toMarshalable(er))
}

func (er *MultiError) MarshalYAML() (any, error) {
	if er.errs == nil {
		return yaml.Marshal(nil)
	}
	return yaml.Marshal(toMarshalable(er))
}
