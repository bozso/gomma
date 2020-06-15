package data

import (
    "fmt"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/utils/params"
)

type File struct {
    DatFile path.ValidFile `json:"datafile"`
    Meta                   `json:"metadata"`
}

func (f *File) SetDataFile(vf path.ValidFile) {
    f.DatFile = vf
}

func (f *File) Set(s string) (err error) {
    return LoadJson(s, f)
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

func (d File) DataPath() path.ValidFile {
    return d.DatFile
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

func LoadJson(s string, val interface{}) (err error) {
    file, err := path.New(s).ToFile().ToValid()
    if err != nil {
        return
    }
    
    err = common.LoadJson(file, val)
    return    
}
