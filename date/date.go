package date

import (
    "fmt"
    "time"
)

type Time struct {
    t time.Time
}

func (t Time) AsTime() (T time.Time) {
    return t.t
}

func FromJson(b []byte, p ParseFmt) (t Time, err error) {
    t.t, err = p.Parse(string(b))
    return
}

type ShortTime struct {
    time.Time
}

func (st *ShortTime) UnmarshalJSON(b []byte) (err error) {
    st.Time.t, err = FromJson(b, Short)
    return
}

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
