package date

import (
	"fmt"
	"time"
)

type Parser interface {
	ParseDate(string) (time.Time, error)
}

type MultiParseError struct {
	Input     string
	LastError error
}

func (m MultiParseError) Error() (s string) {
	return fmt.Sprintf(
		"failed to parse string %s using multiple parsers, last error was: %s", m.Input, m.LastError)
}

func (m MultiParseError) Unwrap() (err error) {
	return m.LastError
}

type MultiParser struct {
	parsers []Parser
}

func (m MultiParser) ParseDate(s string) (t time.Time, err error) {
	for _, parser := range m.parsers {
		t, err = parser.ParseDate(s)
		if err == nil {
			return t, err
		}
	}

	return t, MultiParseError{
		Input:     s,
		LastError: err,
	}
}
