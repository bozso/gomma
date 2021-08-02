package parser

import (
	"fmt"
)

type SplitError struct {
	Line string
	err  error
}

func (s SplitError) Error() (str string) {
	return fmt.Sprintf("while splitting line '%s'", s.Line)
}

func (s SplitError) Unwrap() (err error) {
	return s.err
}
