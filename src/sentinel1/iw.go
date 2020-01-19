package sentinel1

import (
    "../data"
    "../utils"
)

const maxIW = 3

type(  
    S1IW struct {
        data.File
        TOPS_par string
    }

    IWs [maxIW]S1IW
)

func NewIW(dat, par, TOPS_par string) (iw S1IW) {
    iw.Dat = dat
    
    if len(par) == 0 {
        par = dat + ".par"
    }
    
    iw.Params = Params{Par: par, Sep: ":"}
    
    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }

    iw.TOPS_par = Params{Par: TOPS_par, Sep: ":"}

    return
}

func (iw S1IW) Move(dir string) (miw S1IW, err error) {
    ferr := merr.Make("S1IW.Move")
    
    if miw.File, err = iw.File.Move(dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    miw.TOPS_par, err = utils.Move(iw.TOPS_par, dir)
    
    return miw, nil
}
