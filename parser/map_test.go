package parser

import (
	"io"
	"testing"
)

var ColonSetup = Setup{
	Wrapper:  WrapIntoScanner(),
	Splitter: Delimiter(":").AsScanner(),
}

func mapCreate(setup Setup, r io.Reader) (g Getter, err error) {
	g, err = NewMap(setup, r)
	return
}

func TestMapParsing(t *testing.T) {
	cases := []TestCase{
		{
			"a:b\nc:d\n",

			InMemoryStorage{
				"a": "b",
				"c": "d",
			}.ToMap(),
		},
	}

	TestWithSetup{
		Cases: cases,
		Setup: ColonSetup,
	}.Test(t, mapCreate)
}
