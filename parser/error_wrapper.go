package parser

import (
	"fmt"
	"log"
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

type LogErrorWrap struct {
	Wrapper ErrorWrapper
	Logger  log.Logger
}

type WrapFunc func(error) error

func (l LogErrorWrap) LogContext(ctx string, e error, fn WrapFunc) (err error) {
	l.Logger.Printf("wrapping %s error: '%s'", ctx, e)
	e = fn(e)
	l.Logger.Printf("error after wrapping: '%s'", e)

	return err
}

func (l LogErrorWrap) WrapSplitError(line string, e error) (err error) {
	return l.LogContext("split", func(e error) (err error) {
		return l.Wrapper.WrapSplitError(line, e)
	})
}
