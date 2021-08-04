package parser

import (
	"bufio"
	"io"
)

type ReaderWrapper interface {
	WrapReader(io.Reader) (*bufio.Scanner, error)
}

type WrapScanner struct{}

func (WrapScanner) WrapReader(r io.Reader) (b *bufio.Scanner, err error) {
	return bufio.NewScanner(r), nil
}

func WrapIntoScanner() (w WrapScanner) {
	return WrapScanner{}
}

type Setup struct {
	Splitter Splitter
	Wrapper  ReaderWrapper
}

func (s Setup) ParseInto(r io.Reader, setter Setter) (err error) {
	scan, err := s.Wrapper.WrapReader(r)
	if err != nil {
		return
	}

	for scan.Scan() {
		line := scan.Text()

		key, val, err := s.Splitter.SplitLine(line)
		if err != nil {
			return err
		}

		if err = setter.SetParsed(key, val); err != nil {
			return err
		}
	}

	if err := scan.Err(); err != nil && err != io.EOF {
		return err
	}

	return
}
