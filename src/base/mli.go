package base

import (
    "../plot"
    "../data"
)

type MLI struct {
    data.File `json:"DatParFile"`
}

func MLIFromFile(path string) (mli MLI, err error) {
    mli.File, err = data.FromFile(path)
    if err != nil { return; }
    
    err = mli.TypeCheck("MLI", "float", data.Float)
    return
}

func (m MLI) Raster(opt plot.RasArgs) error {
    opt.Mode = plot.Power
    return m.Raster(opt)
}

func (m *MLI) Set(s string) (err error) {
    *m, err = MLIFromFile(s)
    return
}
