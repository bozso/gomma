package date

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type FormatParser interface {
	ParseWithFormat(Format, string) (time.Time, error)
}

type StdFormatParser struct{}

func (StdFormatParser) ParseWithFormat(fmt Format, s string) (t time.Time, err error) {
	return time.Parse(string(fmt), s)
}

func NewStdParser() (s *StdFormatParser) {
	return &StdFormatParser{}
}

type LogFormatParser struct {
	parser FormatParser
	logger *log.Logger
}

func NewLogFormatParser(parser FormatParser, logger *log.Logger) (l LogFormatParser) {
	return LogFormatParser{
		parser: parser,
		logger: logger,
	}
}

func (l LogFormatParser) ParseWithFormat(fmt Format, s string) (t time.Time, err error) {
	l.logger.Debugf("Parsing with format: %s", fmt)
	t, err = l.parser.ParseWithFormat(fmt, s)
	return
}
