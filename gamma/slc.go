package gamma

import (
    "fmt"
)

type SLC struct {
    dataFile
}

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.dataFile, err = NewDataFile(dat, par, Unknown)
    return
}

func (s SLC) TypeStr() string {
    return "SLC"
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

func (s SLC) MakeMLI(opt MLIOpt) (ret MLI, err error) {
    opt.Parse()
    
    tmp := ""
    
    if tmp, err = TmpFileExt("mli"); err != nil {
        //err = Handle(err, "failed to create tmp file")
        return ret, err
    }
    
    if ret, err = NewMLI(tmp, ""); err != nil {
        err = DataCreateErr.Wrap(err, "MLI")
        //err = Handle(err, "failed to create MLI struct")
        return
    }
    
    _, err = multiLook(s.Dat, s.Par, ret.Dat, ret.Par,
                       opt.Looks.Rng, opt.Looks.Azi,
                       opt.Subset.Begin, opt.Subset.Nlines,
                       opt.ScaleExp.Scale, opt.ScaleExp.Exp)
    
    if err != nil {
        err = CmdErr.Wrap(err, "multi_look")
        //err = Handle(err, "multi_look failed")
        return
    }
    
    return ret, nil
}

type (
    SBIOpt struct {
        NormSquintDiff float64
        Looks RngAzi
        InvWeight, Keep  bool
    }
    
    SBIOut struct {
        ifg IFG
        mli MLI
    }
)


var sbiInt = Gamma.Must("SBI_INT")

func (opt *SBIOpt) Default() {
    opt.Looks.Default()
    
    if opt.NormSquintDiff == 0.0 {
        opt.NormSquintDiff = 0.5
    }
}

func (ref SLC) SplitBeamIfg(slave SLC, opt SBIOpt) (ret SBIOut, err error) {
    opt.Default()
    
    tmp := ""
    
    if tmp, err = TmpFile(); err != nil {
        //err = Handle(err, "failed to create tmp file")
        return ret, err
    }
    
    if ret.ifg, err = NewIFG(tmp + ".diff", "", "", "", ""); err != nil {
        err = DataCreateErr.Wrap(err, "IFG")
        //err = Handle(err, "failed to create IFG struct")
        return
    }
    
    if ret.mli, err = NewMLI(tmp + ".mli", ""); err != nil {
        err = DataCreateErr.Wrap(err, "MLI")
        //err = Handle(err, "failed to create MLI struct")
        return
    }
    
    iwflg, cflg := 0, 0
    if opt.InvWeight { iwflg = 1 }
    if opt.Keep { cflg = 1 }
    
    _, err = sbiInt(ref.Dat, ref.Par, slave.Dat, slave.Par,
                    ret.ifg.Dat, ret.ifg.Par, ret.mli.Dat, ret.mli.Par, 
                    opt.NormSquintDiff, opt.Looks.Rng, opt.Looks.Azi,
                    iwflg, cflg)
    
    if err != nil {
        err = CmdErr.Wrap(err, "SBI_INT")
        //err = Handle(err, "SBI_INT failed")
        return
    }
    
    return ret, nil
}

type (
    SSIMode int
    
    SSIOpt struct {
        Hgt, LtFine, OutDir string
        Mode SSIMode
        Keep bool
    }
    
    SSIOut struct {
        Ifg IFG
        Unw dataFile
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


func (d SLC) PlotCmd() string {
    return "SLC"
}

func (s SLC) Raster(opt RasArgs) error {
    err := opt.Parse(s)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return rasslc(opt)
}

type MLI struct {
    dataFile
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.dataFile, err = NewDataFile(dat, par, Unknown)
    return
}

func (m MLI) TypeStr() string {
    return "MLI"
}

func (d MLI) PlotCmd() string {
    return "MLI"
}

func (m MLI) Raster(opt RasArgs) error {
    err := opt.Parse(m)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return raspwr(opt)
}

