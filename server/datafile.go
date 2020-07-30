package server

import (

)

type DataFileConvertFail struct {
    c ConvertFail
    loadedFrom path.ValidFile
}

func (d DataFileConvertFail) Error() (s string) {
    return fmt.Sprintf("%w\nfile was loaded from '%s'",
        r.c, r.loadedFrom)
}

type DataFileType int

const (
    SLC DataFileType = iota
    MLI
    S1SLC
    
)

func (d DataFileType) String() (s string) {
    switch d {
    case 
    
    default:
        s = "Unknown"
    }
}

type tag struct {
    Type DataFileType `json:"type"`
}

type DataFile struct {
    c Convertable
    loadedFrom path.ValidFile
}

func (r Record) WrapFail(c ConvertFail) (d DataFileConvertFail) {
    d.c, d.loadedFrom = c, from
    return
}
