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

    long dateFormat = iota
    short
)

func ParseDate(format dateFormat, str string) (ret time.Time, err error) {
    var form string

    switch format {
    case long:
        form = DateLong
    case short:
        form = DateShort
    default:
        break
    }

    ret, err = time.Parse(form, str)

    if err != nil {
        err = Handle(err, "failed to parse date '%s'", str)
        return
    }

    return ret, nil
}

func NewDate(format dateFormat, start, stop string) (ret date, err error) {
    var _start, _stop time.Time
    
    _start, err = ParseDate(format, start)
    if err != nil {
        err = Handle(err, "failed to parse date: '%s'", start)
        return
    }

    _stop, err = ParseDate(format, stop)
    if err != nil {
        err = Handle(err, "failed to parse date: '%s'", start)
        return
    }

    // TODO: Optional check duration, is it max or min
    delta := _start.Sub(_stop) / 2.0
    ret.center = _stop.Add(delta)

    ret.start = _start
    ret.stop = _stop

    return ret, nil
}

func (self *date) Start() time.Time {
    return self.start
}

func (self *date) Center() time.Time {
    return self.center
}

func (self *date) Stop() time.Time {
    return self.stop
}

func Before(one, two Date) bool {
    return one.Center().Before(two.Center())
}

func date2str(date Date, format dateFormat) string {
    var layout string

    switch format {
    case long:
        layout = DateLong
    case short:
        layout = DateShort
    default:
        break
    }

    return date.Center().Format(layout)
}
