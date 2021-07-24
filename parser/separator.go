package parser

import (
	"fmt"
	"strings"
)

type Splitter interface {
	SplitLine(string) (key, value string, err error)
}

type Delimiter string

func (d Delimiter) SplitLine(str string) (key, value string, err error) {
	split := strings.Split(str, string(d))
	if err = CheckLen(split, 2); err != nil {
		err = &SplitError{
			Line:      str,
			Delimiter: d,
			err:       err,
		}
	}

	return split[0], split[1], err
}

func CheckLen(split []string, expected int) (err error) {
	if l := len(split); l != expected {
		err = &WrongElementNumber{
			Expected: expected,
			Got:      l,
		}
	}

	return
}

type WrongElementNumber struct {
	Expected, Got int
}

func (w WrongElementNumber) Error() (s string) {
	return fmt.Sprintf("expected '%d' elements, but got '%d' elements",
		w.Expected, w.Got)
}
