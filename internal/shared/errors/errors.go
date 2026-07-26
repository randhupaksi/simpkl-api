package errors

import "fmt"

type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}
