package date

import (
	"time"
)

var (
	defaultFormats = FormatsFromStrings(
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	)
	/* This is a bit hacky solution for customizing
	   the way Unmarshaling from json should work.*/
	DefaultParser Parser = defaultFormats
)

type Date struct {
	time.Time
}

func (d Date) UnmarshalJSON(b []byte) (err error) {
	d.Time, err = DefaultParser.ParseDate(string(b))
	return
}
