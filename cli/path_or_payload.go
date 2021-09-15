package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"

	sfs "git.sr.ht/~istvan_bozso/shutil/fs"
)

type Decoder interface {
	Decode(io.Reader, interface{}) error
}

type PathOrPayload struct {
	data string
}

func (p *PathOrPayload) Set(s string) (err error) {
	p.data = s

	return nil
}

type ParseConfig struct {
	fsys    fs.FS
	decoder Decoder
}

func (p ParseConfig) ParseStringPayload(s string, v interface{}) (err error) {
	var r io.Reader

	stat, err := fs.Stat(p.fsys, s)
	if err == nil {
		if stat.IsDir() {
			err = fmt.Errorf(
				"expected path '%s' to be a file not a directory", s)

			return err
		}

		file, err := p.fsys.Open(s)
		if err != nil {
			return err
		}
		defer file.Close()

		r = file
	} else {
		r = strings.NewReader(s)
	}

	return p.decoder.Decode(bufio.NewReader(r), v)
}

func (p ParseConfig) Parse(pp PathOrPayload, v interface{}) (err error) {
	return p.ParseStringPayload(pp.data, v)
}

type JSONCoder struct{}

func (JSONCoder) Decode(r io.Reader, v interface{}) (err error) {
	return json.NewDecoder(r).Decode(v)
}

type Coder int

const (
	CoderJSON Coder = iota
)

func (c Coder) GetDecoder() (d Decoder) {
	switch c {
	case CoderJSON:
		return JSONCoder{}
	default:
		panic("unrecognized coder variant")
	}
}

var DefaultParser = ParseConfig{
	decoder: CoderJSON.GetDecoder(),
	fsys:    sfs.OS(),
}
