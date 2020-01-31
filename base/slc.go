package base

import (
    "fmt"
    
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
    "github.com/bozso/gamma/common"
)

type SLC struct {
    data.File
}

func SLCFromFile(path string) (slc SLC, err error) {
    slc.File, err = data.FromFile(path)
    if err != nil { return; }
    
    err = slc.TypeCheck("SLC", "complex", data.FloatCpx, data.ShortCpx)
    
    return
}

var multiLook = common.Gamma.Must("multi_look")

type (
    // TODO: add loff, nlines
    MLIOpt struct {
        //Subset
        RefTab string
        Looks common.RngAzi
        WindowFlag bool
        plot.ScaleExp
    }
)

func (opt *MLIOpt) Parse() {
    opt.ScaleExp.Parse()
    
    if len(opt.RefTab) == 0 {
        opt.RefTab = "-"
    }
    
    opt.Looks.Default()
}

func (s SLC) MLI(out MLI, opt MLIOpt) (err error) {
    opt.Parse()
    
    _, err = multiLook(s.Dat, s.Par, out.Dat, out.Par,
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


var sbiInt = common.Gamma.Must("SBI_INT")

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
    
    _, err = sbiInt(ref.Dat, ref.Par, slave.Dat, slave.Par,
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

var ssiInt = common.Gamma.Must("SSI_INT")

func (ref SLC) SplitSpectrumIfg(slave SLC, mli MLI, opt SSIOpt) (ret SSIOut, err error) {
    mode := 1
    
    if opt.Mode == IfgUnwrapped {
        mode = 2
    }
    
    cflg := 1
    if opt.Keep { cflg = 0 }
    
    short := common.DateShort
    
    mID, sID := short.Format(ref), short.Format(slave)
    
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    _, err = ssiInt(ref.Dat, ref.Par, mli.Dat, mli.Par, opt.Hgt, opt.LtFine,
                    slave.Dat, slave.Par, mode, mID, sID, ID, opt.OutDir, cflg)
    
    // TODO: figure out the name of the output files
    
    return
}

func (s SLC) Raster(opt plot.RasArgs) error {
    opt.Mode = plot.SingleLook
    return s.Raster(opt)
}

func (slc *SLC) Set(s string) (err error) {
    *slc, err = SLCFromFile(s)
    return
}
