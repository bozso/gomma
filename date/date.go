package date

import (
    "fmt"
    "time"
)

type (
    Dater interface {
        Date() time.Time
    }
    
    ParseFmt string
)


// TODO: these are Sentinel-1 specific, should be moved accordingly
const (
    Short ParseFmt = "20060102"
    Long  ParseFmt = "20060102T150405"
)

func (df ParseFmt) Parse(str string) (t time.Time, err error) {
    if t, err = time.Parse(string(df), str); err != nil {
        err = DateParseError{str, err}
    }

    return
}

func (df ParseFmt) Format(d Dater) (s string) {
    return d.Date().Format(string(df))
}

func (df ParseFmt) ID(one, two Dater) string {
    return fmt.Sprintf("%s_%s", df.Format(one), df.Format(two))
}

type DateParseError struct {
    source string
    err error
}

func (d DateParseError) Error() string {
    return fmt.Sprintf("failed to parse date from string: '%s'", d.source)
}

func (d DateParseError) Unwrap() error {
    return d.err
} 
