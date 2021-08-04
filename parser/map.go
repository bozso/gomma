package parser

import (
	"io"
)

type InMemoryStorage map[string]string

type Map struct {
	data InMemoryStorage
}

func (m Map) GetParsed(key string) (val string, err error) {
	val, ok := m.data[key]
	if !ok {
		err = &MissingKey{
			Key: key,
		}
	}

	return
}

func (m Map) AsMut() (mm *MutMap) {
	return &MutMap{
		wrap: m,
	}
}

type MutMap struct {
	wrap Map
}

func (m MutMap) GetParsed(key string) (val string, err error) {
	return m.wrap.GetParsed(key)
}

func (m *MutMap) SetParsed(key, val string) (err error) {
	m.wrap.data[key] = val
	return nil
}

func NewMap(s Setup, r io.Reader) (m Map, err error) {
	m.data = make(InMemoryStorage)
	err = s.ParseInto(r, m.AsMut())

	return
}
