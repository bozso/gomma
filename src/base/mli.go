package base

import (
    "../plot"
    "../datafile"
)

type MLI struct {
    datafile.File `json:"DatParFile"`
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.DatParFile, err = NewDatParFile(dat, par, "par", Float)
    return
}

func (m MLI) Raster(opt plot.RasArgs) error {
    opt.Mode = Power
    return m.Raster(opt)
}

func (m *MLI) Set(s string) (err error) {
    if err = LoadJson(s, m); err != nil {
        return
    }
    
    return m.TypeCheck("MLI", "float", Float)
}
