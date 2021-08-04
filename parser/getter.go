package parser

import (
	"bufio"
	"io"
)

type Line interface {
	ParseLine(line string) (err error)
}

func ProcInput(r io.Reader, l Line) (err error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		if err = l.ParseLine(s.Text()); err != nil {
			return
		}
	}

	return s.Err()
}

type Map struct {
	data map[string]string
}

type MapCreator struct {
}

func NewMap(r io.Reader) (m Map, err error) {
	m = make(Map)

	return ProcIntut(r, func(line string) (err error) {
	})
}
