package common

import (
    "fmt"
    "time"
)

type (
    date struct {
        start, stop, center time.Time
    }

    Dater interface {
        Date() time.Time
    }

    dateFormat int
)

const (
    dateShort = "20060102"
    dateLong  = "20060102T150405"

    DLong dateFormat = iota
    DShort
)

func ParseDate(format dateFormat, str string) (t time.Time, err error) {
    var (
        ferr = merr.Make("ParseDate")
        form string
    )

    switch format {
    case DLong:
        form = dateLong
    case DShort:
        form = dateShort
    default:
        break
    }

    if t, err = time.Parse(form, str); err != nil {
        err = ferr.WrapFmt(err, "failed to parse date '%s'", str)
        return
    }

    return t, nil
}

func NewDate(format dateFormat, start, stop string) (d date, err error) {
    ferr := merr.Make("NewDate")
    
    _start, err := ParseDate(format, start)
    if err != nil {
        err = ferr.WrapFmt(err, "failed to parse date: '%s'", start)
        return
    }

    _stop, err := ParseDate(format, stop)
    if err != nil {
        err = ferr.WrapFmt(err, "failed to parse date: '%s'", start)
        return
    }

    // TODO: Optional check duration, is it max or min
    delta := _start.Sub(_stop) / 2.0
    d.center = _stop.Add(delta)

    d.start = _start
    d.stop = _stop

    return d, nil
}

func (d date) Start() time.Time {
    return d.start
}

func (d date) Center() time.Time {
    return d.center
}

func (d date) Stop() time.Time {
    return d.stop
}

func Format(t time.Time, format dateFormat) (s string) {
    switch format {
    case DShort:
        s = t.Format(dateShort)
    case DLong:
        s = t.Format(dateLong)
    }
    
    return
}

func ID(one, two Dater, format dateFormat) string {
    return fmt.Sprintf("%s_%s",
        Format(one.Date(), format), Format(two.Date(), format))
}
