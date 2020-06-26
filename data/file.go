package data

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/utils/params"
)

type File struct {
    DatFile path.ValidFile `json:"datafile"`
    Meta                   `json:"metadata"`
}

func (f File) SavePath(ext string) (fp path.File) {
    return f.DatFile.AddExt(ext).ToFile()
}

func (f *File) SetDataFile(vf path.ValidFile) {
    f.DatFile = vf
}

func (f File) TypeCheck(dtypes... Type) (err error) {
    err = f.Meta.TypeCheck(f.DatFile, dtypes...)
    return
}

func (d File) DataPath() path.ValidFile {
    return d.DatFile
}

func (f File) Save() (err error) {
    return common.SaveJson(f)
}

func (d File) SaveWithPath(file path.File) (err error) {
    return common.SaveJsonTo(file, d)
}

func (f File) Move(dir path.Dir) (fm File, err error) {
    fm = f
    fm.DatFile, err = f.DatFile.Move(dir)
    if err != nil {
        return
    }

    return
}

func (d File) Exist() (b bool, err error) {
    b, err = d.DatFile.Exist()
    return
}

const (
    separator = ":"
)

func NewGammaParams(file path.ValidFile) (p params.Params, err error) {
    return params.FromFile(file, separator)
}
