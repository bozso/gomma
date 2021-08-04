package parser

import (
	"io"
)

type InMemoryStorage map[string]string

type Map struct {
	data InMemoryStorage
}

func (m Map) Get(key string) (val string, err error) {
	val, ok := m.data[key]
	if !ok {
		err = &MissingKey{
			Key: key,
		}
	}

	return
}

func (m *Map) SetKeyVal(key, val string) (err error) {
	m.data[key] = val
	return nil
}

func NewMap(s Setup, r io.Reader) (m Map, err error) {
	m.data = make(InMemoryStorage)
	err = s.ParseInto(r, &m)

	return
}
