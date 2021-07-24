package parser

import (
	"fmt"
)

type SplitError struct {
	Delimiter
	Line string
	err  error
}

func (s SplitError) Error() (str string) {
	return fmt.Sprintf("while splitting line '%s' with delimiter '%s'",
		s.Line, s.Delimiter)
}

func (s SplitError) Unwrap() (err error) {
	return s.err
}
