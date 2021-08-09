package parser

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type ErrorWriter interface {
	ErrorWrite(io.Writer, error) (int, error)
}

type ErrorUnwrapper string

func (eu ErrorUnwrapper) ErrorWrite(w io.Writer, e error) (n int, err error) {
	n = 0

	for curr := e; curr != nil; curr = errors.Unwrap(curr) {
		ncurr, err := fmt.Fprintf(w, string(eu), curr)
		if err != nil {
			return n, err
		}

		n += ncurr
	}

	return n, nil
}

func ErrorToString(ew ErrorWriter, e error) (s string, err error) {
	sb := strings.Builder{}
	_, writeErr := ew.ErrorWrite(&sb, e)

	if writeErr != nil {
		return s, writeErr
	}

	return sb.String(), nil
}
