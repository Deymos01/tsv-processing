package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

// ParseError describes a failure to parse a specific row in a TSV file.
type ParseError struct {
	File    string // source TSV filename
	Line    int    // 1-based line number
	Column  string // column name where parsing failed
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error in file %q at line %d, column %q: %s",
		e.File, e.Line, e.Column, e.Message)
}

// WrapError wraps an error with additional context.
func WrapError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}
