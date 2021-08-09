package parser

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type CreateGetter func(Setup, io.Reader) (Getter, error)

type TestCase struct {
	Input    string
	Expected Getter
}

func (tc TestCase) TestErr(s Setup, creator CreateGetter) (err error) {
	g, err := creator(s, bytes.NewBufferString(tc.Input))
	if err != nil {
		return
	}

	for _, key := range tc.Expected.Keys() {
		const tpl = "expected resulting parsed Getter (%#v) to have key '%s'"
		if !g.HasKey(key) {
			return fmt.Errorf(tpl, g, key)
		}
	}

	return
}

func (tc TestCase) Test(t *testing.T, s Setup, creator CreateGetter) {
	if err := tc.TestErr(s, creator); err != nil {
		t.Errorf("error while testing with input '%s' %s", tc.Input, err)
	}
}

type TestWithSetup struct {
	Cases []TestCase
	Setup Setup
}

func (ts TestWithSetup) TestErr(creator CreateGetter) (err error) {
	for _, testCase := range ts.Cases {
		if err = testCase.TestErr(ts.Setup, creator); err != nil {
			return err
		}
	}

	return nil
}

type ErrorFn func(error)

func (ts TestWithSetup) ForEach(fn ErrorFn, creator CreateGetter) {
	for _, testCase := range ts.Cases {
		fn(testCase.TestErr(ts.Setup, creator))
	}
}

func (ts TestWithSetup) Test(t *testing.T, creator CreateGetter) {
	for _, testCase := range ts.Cases {
		testCase.Test(t, ts.Setup, creator)
	}
}
