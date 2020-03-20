package date

import (
    "fmt"
    "time"
)

type ParseFmt string

const (
    Short ParseFmt = "20060102"
    Long  ParseFmt = "20060102T150405"
)

func (df ParseFmt) Parse(str string) (t time.Time, err error) {
    if t, err = time.Parse(string(df), str); err != nil {
        err = ParseError{str, err}
    }

    return
}

func (df ParseFmt) Format(d Dater) (s string) {
    return d.Date().Format(string(df))
}

func (df ParseFmt) ID(one, two Dater) string {
    return fmt.Sprintf("%s_%s", df.Format(one), df.Format(two))
}

type ParseError struct {
    source string
    err error
}

func (p ParseError) Error() string {
    return fmt.Sprintf("failed to parse date from string: '%s'", p.source)
}

func (p ParseError) Unwrap() error {
    return p.err
} 
