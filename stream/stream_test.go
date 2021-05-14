package stream

import (
	"os"
	"reflect"
	"testing"
)

func TestStdin(t *testing.T) {
	var in In
	if err := in.Set("stdin"); err != nil {
		t.Fatalf("error setting up stream.In: %s", err)
	}

	if !reflect.DeepEqual(in.r, os.Stdin) {
		t.Fatalf("reader was not set to stdin")
	}
}

func TestStdout(t *testing.T) {
	var out Out
	if err := out.Set("stdout"); err != nil {
		t.Fatalf("error setting up stream.Out: %s", err)
	}

	if !reflect.DeepEqual(out.w, os.Stdout) {
		t.Fatalf("reader was not set to stdout")
	}
}
