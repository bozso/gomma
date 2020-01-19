package common

import (
    "fmt"
    "time"
)

type (
    Dater interface {
        Date() time.Time
    }
    
    DateFormat string
)


// TODO: these are Sentinel-1 specific, should be moved accordingly
const (
    DateShort DateFormat = "20060102"
    DateLong  DateFormat = "20060102T150405"
)

func (df DateFormat) ParseDate(str string) (t time.Time, err error) {
    if t, err = time.Parse(string(df), str); err != nil {
        err = DateParseError{str, err}
    }

    return
}

func (df DateFormat) Format(d Dater) (s string) {
    return d.Date().Format(string(df))
}

func (df DateFormat) ID(one, two Dater) string {
    return fmt.Sprintf("%s_%s", df.Format(one), df.Format(two))
}
