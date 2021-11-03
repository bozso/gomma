package date

import (
	"fmt"
	"testing"
	"time"
)

type ParsedNotEqual struct {
	Expected time.Time
	Got      time.Time
}

func (p ParsedNotEqual) Error() (s string) {
	return fmt.Sprintf("expected parsed time to be '%s', got: '%s'",
		p.Expected, p.Got)
}

type ParserTest struct {
	Expected []time.Time
	ToParse  []string
}

func (p ParserTest) TestParser(parser Parser) (err error) {
	for ii, expected := range p.Expected {
		current := p.ToParse[ii]
		result, err := parser.ParseDate(current)
		if err != nil {
			return ParseError{
				source: current,
				err:    err,
			}
		}

		if result.Equal(expected) {
			return ParsedNotEqual{
				Expected: expected,
				Got:      result,
			}
		}
	}

	return
}

func (p ParserTest) Test(t *testing.T, parser Parser) {
	err := p.TestParser(parser)
	if err != nil {
		t.Errorf("testing of parser failed with error: %s", err)
	}
}

func TestDefaultParser(t *testing.T) {
	ParserTest{
		Expected: []time.Time{
			time.Date(2006, 12, 1, 0, 0, 1, 0, time.UTC),
			time.Date(2006, 12, 1, 0, 0, 1, 0, time.UTC),
		},
		ToParse: []string{
			"2006-12-01",
			"2006 12 01",
		},
	}.Test(t, DefaultParser)
}
