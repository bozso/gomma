package data

import (
    "fmt"
    "time"
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/utils/params"
)

type File struct {
    DatFile path.ValidFile `json:"datafile"`
    Meta
}

func (f *File) Set(s string) (err error) {
    file, err := path.New(s).ToFile().ToValid()
    if err != nil {
        return
    }
    
    err = common.LoadJson(file, f)
    return
}

func (f File) TypeCheck(dtypes... Type) (err error) {
    err = f.Meta.TypeCheck(f.DatFile, dtypes...)
    return
}

func (f File) JsonName() (file path.File) {
    // it is okay for the path to not exist
    file = path.New(fmt.Sprintf("%s.json", f.DatFile)).ToFile()
    return
}

func (d File) Rng() int {
    return d.Ra.Rng
}

func (d File) Azi() int {
    return d.Ra.Azi
}

func (d File) DataPath() path.ValidFile {
    return d.DatFile
}

func (d File) Date() time.Time {
    return d.Time
}

func (d File) DataType() Type {
    return d.Dtype
}

func (d File) Save() (err error) {
    return d.SaveWithPath(d.JsonName())
}

func (d File) SaveWithPath(file path.File) (err error) {
    return common.SaveJson(file, d)
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
