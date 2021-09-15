package batch

import (
	"fmt"
)

type Error struct {
	Operation string
	err       error
}

func (e Error) Error() (s string) {
	return fmt.Sprintf("while carrying out operation '%s'", e.Operation)
}

func (e Error) Unwrap() (err error) {
	return e.err
}

type Func func() error

func WrapErr(op string, fn Func) (err error) {
	err = fn()
	if err != nil {
		return &Error{
			Operation: op,
			err:       err,
		}
	}

	return err
}
