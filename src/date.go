package gamma

import (
    //"fmt"
    "time"
)

type (
    date struct {
        start, stop, center time.Time
    }

    Date interface {
        Start() time.Time
        Stop() time.Time
        Center() time.Time
    }
    dateFormat int
)

const (
    DateShort = "20060102"
    DateLong  = "20060102T150405"

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
        form = DateLong
    case DShort:
        form = DateShort
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
    var (
        ferr = merr.Make("NewDate")
        _start, _stop time.Time
    )
    
    if _start, err = ParseDate(format, start); err != nil {
        err = ferr.WrapFmt(err, "failed to parse date: '%s'", start)
        return
    }

    _stop, err = ParseDate(format, stop)
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

func (d *date) Start() time.Time {
    return d.start
}

func (d *date) Center() time.Time {
    return d.center
}

func (d *date) Stop() time.Time {
    return d.stop
}

func Before(one, two Date) bool {
    return one.Center().Before(two.Center())
}

func date2str(date Date, format dateFormat) string {
    var layout string

    switch format {
    case DLong:
        layout = DateLong
    case DShort:
        layout = DateShort
    default:
        break
    }

    return date.Center().Format(layout)
}
