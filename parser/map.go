package parser

import (
	"io"
)

type empty struct{}

type InMemoryStorage map[string]string

type Map struct {
	keys []string
	data InMemoryStorage
}

func (m Map) Keys() (keys []string) {
	return m.keys
}

func (m Map) HasKey(key string) (hasKey bool) {
	_, hasKey = m.data[key]
	return
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
		Map: m,
	}
}

type MutMap struct {
	Map
}

func (m *MutMap) SetParsed(key, val string) (err error) {
	m.keys = append(m.keys, key)
	m.data[key] = val
	return nil
}

func NewMap(s Setup, r io.Reader) (m Map, err error) {
	m.data = make(InMemoryStorage)
	err = s.ParseInto(r, m.AsMut())

	return
}
