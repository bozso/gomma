package data

import (
	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/common"
	"github.com/bozso/gomma/utils/params"
)

type File struct {
	Meta
}

func (f File) MetaData() Meta {
	return f.Meta
}

func (f File) SavePath(ext string) (l path.Like) {
	return f.DataFile.AddExtension(ext)
}

func (f *File) SetDataFile(vf path.ValidFile) {
	f.DataFile = vf
}

func (f File) TypeCheck(dtypes ...Type) (err error) {
	err = f.Meta.TypeCheck(f.DataFile, dtypes...)
	return
}

func (d File) DataPath() path.ValidFile {
	return d.DataFile
}

func (f File) Save() (err error) {
	return common.SaveJson(f)
}

func (d File) SaveWithPath(file path.Like) (err error) {
	return common.SaveJsonTo(file, d)
}

func (f File) Move(dir path.Dir) (fm File, err error) {
	fm = f
	fm.DataFile, err = f.DataFile.Move(dir)
	if err != nil {
		return
	}

	return
}

func (d File) Exists() (b bool, err error) {
	b, err = d.DataFile.Exists()
	return
}

const (
	separator = ":"
)

func NewGammaParams(file path.ValidFile) (p params.Params, err error) {
	return params.FromFile(file, separator)
}
