package date

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type LogParser struct {
	logger *log.Logger
	parser Parser
}

func NewLogger(parser Parser, logger *log.Logger) (l LogParser) {
	return LogParser{
		parser: parser,
		logger: logger,
	}
}

func (l LogParser) ParseDate(s string) (t time.Time, err error) {
	l.logger.Debugf("Parsing date string: %s", s)
	t, err = l.parser.ParseDate(s)
	if err != nil {
		l.logger.Errorf("error while parsing: %s", err)
	}
	return
}
