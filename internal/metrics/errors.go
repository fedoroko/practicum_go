package metrics

import (
	"fmt"
)

var InvalidType *invalidTypeError

type invalidTypeError struct {
	Type string
}

func (e *invalidTypeError) Error() string {
	return fmt.Sprintf("Invalid type: %v", e.Type)
}

func ThrowInvalidTypeError(t string) error {
	return &invalidTypeError{Type: t}
}

var InvalidHash *invalidHashError

type invalidHashError struct{}

func (e *invalidHashError) Error() string {
	return "Invalid hash"
}

func ThrowInvalidHashError() error {
	return &invalidHashError{}
}
