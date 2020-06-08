package slc

import (
    "fmt"
    
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/mli"
)

type SLC struct {
    data.FileWithPar `json:"SLC"`
}

func (s SLC) Validate() (err error) {
    return s.EnsureComplex()
}

var multiLook = common.Must("multi_look")

func (s SLC) MLI(out mli.MLI, opt mli.Options) (err error) {
    opt.Parse()
    
    _, err = multiLook.Call(s.DatFile, s.ParFile, out.DatFile, out.ParFile,
                       opt.Looks.Rng, opt.Looks.Azi,
                       //opt.Subset.RngOffset, opt.Subset.RngWidth,
                       opt.ScaleExp.Scale, opt.ScaleExp.Exp)
    
    return
}

type (
    SBIOpt struct {
        NormSquintDiff float64 `cli:"n,nsquint" dft:"0.5"`
        InvWeight      bool    `cli:"i,invw"`
        Keep           bool    `cli:"k,keep"`
        Looks          common.RngAzi  `cli:"L,looks"`
        Ifg            string  `cli:"ifg"`
        Mli            string  `cli:"mli"`
    }
)


var sbiInt = common.Must("SBI_INT")

func (opt *SBIOpt) Default() {
    opt.Looks.Default()
    
    if opt.NormSquintDiff == 0.0 {
        opt.NormSquintDiff = 0.5
    }
}

func (ref SLC) SplitBeamIfg(slave SLC, opt SBIOpt) (err error) {
    opt.Default()
    
    iwflg, cflg := 0, 0
    if opt.InvWeight { iwflg = 1 }
    if opt.Keep { cflg = 1 }
    
    _, err = sbiInt.Call(ref.DatFile, ref.ParFile,
                    slave.DatFile, slave.ParFile,
                    opt.Ifg, opt.Ifg + ".off", opt.Mli, opt.Mli + ".par", 
                    opt.NormSquintDiff, opt.Looks.Rng, opt.Looks.Azi,
                    iwflg, cflg)
    
    return
}

type (
    SSIMode int
    
    SSIOpt struct {
        Hgt    string  `cli:"h,hgt"`
        LtFine string  `cli:"l,lookup"`
        OutDir string  `cli:"o,out" dft:"."`
        Keep   bool    `cli:"k,keep"`
        Mode   SSIMode `cli:"sm,ssiMode"`
    }
    
    SSIOut struct {
        //Ifg IFG
        Unw data.File
    }
)

const (
    Ifg           SSIMode = iota
    IfgUnwrapped
)

var ssiInt = common.Must("SSI_INT")

func (ref SLC) SplitSpectrumIfg(slave SLC, mli mli.MLI, opt SSIOpt) (ret SSIOut, err error) {
    mode := 1
    
    if opt.Mode == IfgUnwrapped {
        mode = 2
    }
    
    cflg := 1
    if opt.Keep { cflg = 0 }
    
    mID, sID := date.Short.Format(ref), date.Short.Format(slave)
    
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    _, err = ssiInt.Call(ref.DatFile, ref.ParFile, mli.DatFile, mli.ParFile,
        opt.Hgt, opt.LtFine, slave.DatFile, slave.ParFile, mode,
        mID, sID, ID, opt.OutDir, cflg)
    
    // TODO: figure out the name of the output files
    
    return
}

func (s SLC) Raster(opt plot.RasArgs) error {
    opt.Mode = plot.SingleLook
    return s.Raster(opt)
}