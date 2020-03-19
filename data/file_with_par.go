package data

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/common"
)

type FileWithPar struct {
    ParFile path.ValidFile
    File
}

func (f *FileWithPar) Set(s string) (err error) {
    file, err := path.New(s).ToFile().ToValid()
    if err != nil {
        return
    }
    
    err = common.LoadJson(file, f)
    return
}
