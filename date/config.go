package date

import (
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	defaultFormats = FormatsFromStrings(
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	)

	StdParser = NewStdParser()
	logger    = log.StandardLogger()

	debugParser         = NewLogFormatParser(StdParser, logger)
	DefaultFormatParser = StdParser

	/* This is a bit hacky solution for customizing
	   the way Unmarshaling from json should work.
	*/

	/* TODO: add parser for gamma format.*/
	DefaultParser = NewLogger(
		defaultFormats.Ref(debugParser),
		logger,
	)
)

type Date struct {
	time.Time
}

func (d Date) UnmarshalJSON(b []byte) (err error) {
	d.Time, err = DefaultParser.ParseDate(string(b))
	return
}
