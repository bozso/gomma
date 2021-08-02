package parser

import ()

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
