package date

import (
    "fmt"
    "time"
)

type ToTime interface {
    AsTime() (time.Time)
}

type OptionalTime struct {
    set bool
    time.Time
}

func (ot *OptionalTime) Parse(b []byte, p ParseFmt) (err error) {
    if len(b) == 0 {
        ot.set = false
    }
    
    ot.Time, err = p.Parse(string(b))
    if err == nil {
        ot.set = true
    }
    return
}

func (ot OptionalTime) IsSet() (b bool) {
    return ot.set
}

type ShortTime struct {
    OptionalTime
}

func (st *ShortTime) UnmarshalJSON(b []byte) (err error) {
    return st.OptionalTime.Parse(b, Short)
}

type LongTime struct {
    OptionalTime
}

func (lt *LongTime) UnmarshalJSON(b []byte) (err error) {
    return lt.OptionalTime.Parse(b, Long)
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
