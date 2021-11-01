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

type Format string

func (f Format) ParseDate(s string) (t time.Time, err error) {
	return time.Parse(string(f), s)
}

type Formats struct {
	formats []Format
}

func FormatsFromStrings(strings ...string) (f Formats) {
	l := len(strings)
	f.formats = make([]Format, l)
	for ii, s := range strings {
		f.formats[ii] = Format(s)
	}
	return
}

func (f Formats) ParseDate(s string) (t time.Time, err error) {
	for _, format := range f.formats {
		t, err = format.ParseDate(s)
		if err == nil {
			return t, err
		}
	}

	return t, MultiParseError{
		Input:     s,
		LastError: err,
	}
}
