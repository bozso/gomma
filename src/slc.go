package gamma

import (
    "fmt"
)

type SLC struct {
    DatParFile `json:"DatParFile"`
}

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.DatParFile, err = NewDatParFile(dat, par, "par", FloatCpx)
    return
}


var multiLook = Gamma.Must("multi_look")

type (
    // TODO: add loff, nlines
    MLIOpt struct {
        Subset
        refTab string
        Looks RngAzi
        windowFlag bool
        ScaleExp
    }
)

func (opt *MLIOpt) Parse() {
    opt.ScaleExp.Parse()
    
    if len(opt.refTab) == 0 {
        opt.refTab = "-"
    }
    
    opt.Looks.Default()
}

func (s SLC) MLI(out MLI, opt MLIOpt) (err error) {
    opt.Parse()
    
    _, err = multiLook(s.Dat, s.Par, out.Dat, out.Par,
                       opt.Looks.Rng, opt.Looks.Azi,
                       opt.Subset.RngOffset, opt.Subset.RngWidth,
                       opt.ScaleExp.Scale, opt.ScaleExp.Exp)
    
    if err != nil {
        return merr.Make("SLC.MLI").Wrap(err)
    }
    
    return nil
}

type (
    SBIOpt struct {
        NormSquintDiff float64 `cli:"n,nsquint" dft:"0.5"`
        InvWeight      bool    `cli:"i,invw"`
        Keep           bool    `cli:"k,keep"`
        Looks          RngAzi  `cli:"L,looks"`
        Ifg            string  `cli:"ifg"`
        Mli            string  `cli:"mli"`
    }
)


var sbiInt = Gamma.Must("SBI_INT")

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
    
    if err != nil {
        err = CmdErr.Wrap(err, "SBI_INT")
        //err = Handle(err, "SBI_INT failed")
        return
    }
    
    return nil
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
        Ifg IFG
        Unw DatFile
    }
)

const (
    Ifg           SSIMode = iota
    IfgUnwrapped
)

var ssiInt = Gamma.Must("SSI_INT")

func (ref SLC) SplitSpectrumIfg(slave SLC, mli MLI, opt SSIOpt) (ret SSIOut, err error) {
    mode := 1
    
    if opt.Mode == IfgUnwrapped {
        mode = 2
    }
    
    cflg := 1
    if opt.Keep { cflg = 0 }
    
    mID, sID := ref.Format(DateShort), slave.Format(DateShort)
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    _, err = ssiInt(ref.Dat, ref.Par, mli.Dat, mli.Par, opt.Hgt, opt.LtFine,
                    slave.Dat, slave.Par, mode, mID, sID, ID, opt.OutDir, cflg)
    
    if err != nil {
        err = CmdErr.Wrap(err, "SSI_INT")
        //err = Handle(err, "SSI_INT failed")
        return
    }
    
    // TODO: figure out the name of the output files
    
    return ret, nil
}

func (s SLC) Raster(opt RasArgs) error {
    opt.Mode = SingleLook
    return s.Raster(opt)
}

type MLI struct {
    DatParFile `json:"DatParFile"`
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.DatParFile, err = NewDatParFile(dat, par, "par", Float)
    return
}

func (m MLI) Raster(opt RasArgs) error {
    opt.Mode = Power
    return m.Raster(opt)
}


func (slc *SLC) Set(s string) (err error) {
    if err = LoadJson(s, slc); err != nil {
        return
    }
    
    return slc.TypeCheck("SLC", "complex", FloatCpx, ShortCpx)
}

func (m *MLI) Set(s string) (err error) {
    if err = LoadJson(s, m); err != nil {
        return
    }
    
    return m.TypeCheck("MLI", "float", Float)
}
