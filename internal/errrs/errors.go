package errrs

import (
	"fmt"
)

// InvalidType Стоит ли выносить ошибки в отдельный пакет?
// Тут очевидная проблема с неймингом пакета.
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
