package parser

import (
	"fmt"
	"io"
	"testing"
)

var ColonSetup = Setup{
	Wrapper: WrapIntoScanner(),
	Splitter: SplitWrapErr{
		Splitter: Delimiter(":"),
		Wrapper:  ErrorWrapperSimple.New(),
	},
}

func mapCreate(setup Setup, r io.Reader) (g Getter, err error) {
	g, err = NewMap(setup, r)
	return
}

func TestMapParsing(t *testing.T) {
	const errTpl ErrorUnwrapper = "%s "
	fmt.Printf("%#v\n", ColonSetup)

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
	}.ForEach(func(err error) {
		s, err := ErrorToString(errTpl, err)
		PanicOnErr(err)

		if err != nil {
			t.Errorf("%s", s)
		}
	}, mapCreate)
}
