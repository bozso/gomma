package gamma

import (
    "fmt"
)

type SLC struct {
    DatParFile
}

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.DatParFile, err = NewDatParFile(dat, par, "par", FloatCpx)
    return
}

func TmpSLC() (ret SLC, err error) {
    ret.DatParFile, err = TmpDatParFile("slc", "par", FloatCpx)
    return
}

func (s *SLC) FromJson(m JSONMap) (err error) {
    if err = s.DatParFile.FromJson(m); err != nil {
        return
    }
    
    if s.DType != ShortCpx && s.DType != FloatCpx {
        err = TypeMismatchError{ftype:"SLC", expected:"complex", DType:s.DType}
        return
    }
    return nil
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
    
    if ret, err = TmpMLI(); err != nil {
        return
    }
    
    _, err = multiLook(s.Dat, s.Par, ret.Dat, ret.Par,
                       opt.Looks.Rng, opt.Looks.Azi,
                       opt.Subset.RngOffset, opt.Subset.RngWidth,
                       opt.ScaleExp.Scale, opt.ScaleExp.Exp)
    
    if err != nil {
        return
    }
    
    return ret, nil
}

type (
    SBIOpt struct {
        NormSquintDiff float64 `name:"nsquint" default:"0.5"`
        InvWeight      bool    `name:"invw"`
        Keep           bool    `name:"keep"`
        Looks          RngAzi
    }
    
    SBIOut struct {
        Ifg IFG
        Mli MLI
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
    
    if ret.Ifg, err = TmpIFG(); err != nil {
        return
    }
    
    if ret.Mli, err = TmpMLI(); err != nil {
        return
    }
    
    iwflg, cflg := 0, 0
    if opt.InvWeight { iwflg = 1 }
    if opt.Keep { cflg = 1 }
    
    _, err = sbiInt(ref.Dat, ref.Par, slave.Dat, slave.Par,
                    ret.Ifg.Dat, ret.Ifg.Par, ret.Mli.Dat, ret.Mli.Par, 
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
        Hgt    string `name:"" default:""`
        LtFine string `name:"lookup" default:""`
        OutDir string `name:"out" default:"."`
        Keep bool     `name:"keep"`
        Mode SSIMode
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
    DatParFile
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.DatParFile, err = NewDatParFile(dat, par, "par", Float)
    return
}

func TmpMLI() (ret MLI, err error) {
    ret.DatParFile, err = TmpDatParFile("mli", "par", FloatCpx)
    return
}

func (M *MLI) FromJson(m JSONMap) (err error) {
    if err = M.DatParFile.FromJson(m); err != nil {
        return
    }
    
    if M.DType != Float {
        err = TypeMismatchError{ftype:"MLI", expected:"float", DType:M.DType}
        return
    }
    return nil
}


func (m MLI) Raster(opt RasArgs) error {
    opt.Mode = Power
    return m.Raster(opt)
}

