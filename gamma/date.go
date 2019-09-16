package gamma

import (
	"fmt"
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

func ParseDate(format dateFormat, str string) (time.Time, error) {
	var form string

	switch format {
	case long:
		form = DateLong
	case short:
		form = DateShort
	default:
		break
	}

	ret, err := time.Parse(form, str)

	if err != nil {
		return time.Time{},
			fmt.Errorf("In ParseDate: Failed to parse date: %s!\nError: %w",
				str, err)
	}

	return ret, nil
}

func NewDate(format dateFormat, start, stop string) (date, error) {
	self := date{}
	handle := Handler("NewDate")

	_start, err := ParseDate(format, start)
	if err != nil {
		return self, handle(err, "Could not parse date: '%s'", start)
	}

	_stop, err := ParseDate(format, stop)
	if err != nil {
		return self, handle(err, "Could not parse date: '%s'", start)
	}

	// TODO: Optional check duration, is it max or min
	delta := _start.Sub(_stop) / 2.0
	self.center = _stop.Add(delta)

	self.start = _start
	self.stop = _stop

	return self, nil
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
	return one.Start().Before(two.Start())
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
