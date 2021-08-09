package parser

import (
	"fmt"
	"strings"
)

type Splitter interface {
	SplitLine(string) (key, value string, err error)
}

type SplitWrapErr struct {
	Splitter Splitter
	Wrapper  ErrorWrapper
}

func (s SplitWrapErr) SplitLine(str string) (key, value string, err error) {
	key, value, err = s.Splitter.SplitLine(str)
	s.Wrapper.WrapSplitError(str, err)

	return
}

type Delimiter string

func (d Delimiter) SplitLine(str string) (key, value string, err error) {
	split := strings.Split(str, string(d))
	if err = CheckLen(split, 2); err != nil {
		err = &SplitError{
			Line: str,
			err:  err,
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
	Delimiter
	Expected, Got int
}

func (w WrongElementNumber) Error() (s string) {
	return fmt.Sprintf("expected '%d' elements, but got '%d' elements",
		w.Expected, w.Got)
}

type Scanner struct {
	template string
}

func (d Delimiter) AsScanner() (s Scanner) {
	return Scanner{
		template: fmt.Sprintf("%%s%s%%s", d),
	}
}

func (s Scanner) SplitLine(str string) (key, value string, err error) {
	_, err = fmt.Scanf(str, s.template, &key, &value)
	return key, value, err
}
