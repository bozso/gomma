package cli

import (
	"io"
	"io/fs"
)

type Decoder interface {
	Decode(io.Reader, interface{})
}

type PathOrPayload struct {
	data string
}

func (p *PathOrPayload) Set(s string) (err error) {
	p.data = s
}

type ParseConfig struct {
	fsys    fs.FS
	decoder Decoder
}

func (p ParseConfig) ParseInto(s string, v interface{}) (err error) {
	f, err := fs.Stat(cfg.fsys, p.data)
	if err == nil {

	}
}
