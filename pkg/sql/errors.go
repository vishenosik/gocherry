package sql

import "github.com/pkg/errors"

var (
	// exists already
	ErrAlreadyExists = errors.New("exists already")
)
