package stream

import (
	"fmt"
	"os"
	"strings"

	"github.com/bozso/gotoolbox/path"
)

func TrimJSON(b []byte) (s string) {
	s = strings.Trim(string(b), "\"")

	return strings.Trim(s, " ")
}

func open(s string) (f *os.File, err error) {
	vf, err := path.New(s).ToValidFile()
	if err != nil {
		return
	}

	f, err = vf.Open()
	return
}

var names = struct {
	out, in string
}{
	out: "stdout",
	in:  "stdin",
}

type Type int

const (
	InStream Type = iota
	OutStream
)

func (t Type) String() (s string) {
	switch t {
	case InStream:
		s = "input"
	case OutStream:
		s = "output"
	default:
		s = "unknown"
	}

	return
}

type Mode int

const (
	PathMode Mode = iota
	StdinMode
	StdoutMode
)

func (m Mode) String() (s string) {
	switch m {
	case PathMode:
		s = "filepath"
	case StdinMode:
		s = "stdin"
	case StdoutMode:
		s = "stdout"
	default:
		s = "unknown"
	}

	return
}

type Meta struct {
	StreamType Type
	Mode       Mode
}

func (m Meta) Describe() (s string) {
	return fmt.Sprintf("stream type: %s in mode: %s", m.StreamType, m.Mode)
}

func InMeta(mode Mode) (m Meta) {
	return Meta{
		StreamType: InStream,
		Mode:       mode,
	}
}

func StdinMeta() (m Meta) {
	return InMeta(StdinMode)
}

func OutMeta(mode Mode) (m Meta) {
	return Meta{
		StreamType: OutStream,
		Mode:       mode,
	}
}

func StdoutMeta() (m Meta) {
	return OutMeta(StdoutMode)
}

type MismatchedStreamType struct {
	Expected Type
	Got      Meta
}

func (m MismatchedStreamType) Error() (s string) {
	return fmt.Sprintf("expected stream type %s got %s", m.Expected,
		m.Got.Describe())
}
