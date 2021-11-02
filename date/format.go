package date

import (
	"time"
)

type Format string

func (f Format) Ref(parser FormatParser) (fr FormatRef) {
	return FormatRef{
		parser: parser,
		format: f,
	}
}

type FormatRef struct {
	parser FormatParser
	format Format
}

func (f FormatRef) ParseDate(s string) (t time.Time, err error) {
	return f.parser.ParseWithFormat(f.format, s)
}

type Formats []Format

func FormatsFromStrings(strings ...string) (f Formats) {
	l := len(strings)
	f = make(Formats, l)
	for ii, s := range strings {
		f[ii] = Format(s)
	}
	return
}

func (f Formats) Ref(parser FormatParser) (fr FormatRefs) {
	return FormatRefs{
		parser:  parser,
		formats: f,
	}
}

type FormatRefs struct {
	parser  FormatParser
	formats Formats
}

func (f FormatRefs) ParseDate(s string) (t time.Time, err error) {
	for _, format := range f.formats {
		t, err = f.parser.ParseWithFormat(format, s)
		if err == nil {
			return t, err
		}
	}

	return t, MultiParseError{
		Input:     s,
		LastError: err,
	}
}
