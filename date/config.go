package date

import (
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	defaultFormats = FormatsFromStrings(
		"2006 01 02",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	)

	StdParser   = NewStdParser()
	Logger      = log.StandardLogger()
	DebugLogger = func() (l *log.Logger) {
		l = log.StandardLogger()
		l.SetLevel(log.DebugLevel)
		return
	}()

	debugParser         = NewLogFormatParser(StdParser, DebugLogger)
	DefaultFormatParser = StdParser

	/* This is a bit hacky solution for customizing
	   the way Unmarshaling from json should work.
	*/

	/* TODO: add parser for gamma format.*/
	DebugParser = NewLogger(
		defaultFormats.Ref(debugParser),
		Logger,
	)

	DefaultParser = NewLogger(
		defaultFormats.Ref(StdParser),
		Logger,
	)
	// parser = DebugParser
	parser Parser = DefaultParser
)

type Date struct {
	time.Time
}

func (Date) SetParser(dateParser Parser) {
    parser = dateParser
}

func (d Date) UnmarshalJSON(b []byte) (err error) {
	d.Time, err = parser.ParseDate(string(b))
	return
}
