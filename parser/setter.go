package parser

import (
	"bufio"
	"io"
)

type ReaderWrapper interface {
	WrapReader(io.Reader) (bufio.Scanner, error)
}

type Setter interface {
	SetKeyVal(key, value string) error
}

type ParseSetup struct {
	Splitter Splitter
	Wrapper  ReaderWrapper
}

func (p ParseSetup) ParseInto(r io.Reader, setter Setter) (err error) {
	s, err := p.Wrapper.WrapReader(r)
	if err != nil {
		return
	}

	for s.Scan() {
		line := s.Text()

		key, val, err := p.Splitter.SplitLine(line)
		if err != nil {
			return err
		}

		if err = setter.SetKeyVal(key, val); err != nil {
			return err
		}
	}

	if err := s.Err(); err != nil && err != io.EOF {
		return err
	}

	return
}
