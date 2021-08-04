package parser

import (
	"fmt"
)

type ErrorWrapper interface {
	WrapSplitError(line string, err error) error
	WrapParseError(line string, mode Mode, err error) error
}

type SimpleErrorWrapper struct{}

func (SimpleErrorWrapper) WrapSplitError(line string, err error) (e error) {
	e = err
	if err != nil {
		e = &SplitError{
			Line: line,
		}
	}

	return
}

func (SimpleErrorWrapper) WrapParseError(line string, mode Mode, err error) (e error) {
	e = err
	if err != nil {
		e = &Error{
			Line: line,
			Mode: mode,
			err:  err,
		}
	}

	return
}

func DefaultErrorWrapper() (s SimpleErrorWrapper) {
	return SimpleErrorWrapper{}
}

type Error struct {
	Line string
	Mode Mode
	err  error
}

func (e Error) Error() (s string) {
	s = fmt.Sprintf("while parsing line '%s' into an %s",
		e.Line, e.Mode.String())

	return
}

func (e Error) Unwrap() (err error) {
	return e.err
}
