package interferogram

import (
    "../datafile"
    "../plot"
)

type File struct {
    datafile.File           `json:"DatParFile"`
    DiffPar   string        `json:"diffparfile"`
    Quality   string        `json:"quality"`
    SimUnwrap string        `json:"simulated_unwrapped"`
    DeltaT    time.Duration `json:"-"`
}

func New(dat, off, diffpar string) (ifg IFG, err error) {
    var ferr = merr.Make("NewIFG")

    if ifg.DatParFile, err = NewDatParFile(dat, off, "off", FloatCpx);
       err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if len(diffpar) == 0 {
        diffpar = dat + ".diff_par"
    }
    
    ifg.DiffPar = NewGammaParam(diffpar)
    
    return
}

func (i File) Move(dir string) (im File, err error) {
    var ferr = merr.Make("IFG.Move")
    
    if im.DatParFile, err = i.DatParFile.Move(dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if im.DiffPar.Par, err = Move(i.DiffPar.Par, dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    im.DiffPar.Sep = ":"
    
    if len(i.SimUnwrap) > 0 {
        if im.SimUnwrap, err = Move(i.SimUnwrap, dir); err != nil {
            err = ferr.Wrap(err)
            return
        }
    }
    
    if len(i.Quality) > 0 {
        if im.Quality, err = Move(i.Quality, dir); err != nil {
            err = ferr.Wrap(err)
            return
        }
    }
    
    return
}

func (i *File) Set(s string) (err error) {
    var ferr = merr.Make("IFG.Decode")
    
    if err = LoadJson(s, i); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = i.TypeCheck("IFG", "complex", ShortCpx, FloatCpx); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

func FromSLC(slc1, slc2, ref *base.SLC, out File, opt IfgOpt) (err error) {
    var ferr = merr.Make("FromSLC")
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    par1, par2 := slc1.Par, slc2.Par
    
    // TODO: check arguments!
    _, err = createOffset(par1, par2, out.Par, opt.algo, rng, azi, inter)
    
    if err != nil {
        return ferr.WrapFmt(err, "failed to create offset table")
        
    }
    
    slcRefPar := "-"
    
    if ref != nil {
        slcRefPar = ref.Par
    }
    
    _, err = phaseSimOrb(par1, par2, out.Par, opt.hgt, out.SimUnwrap,
        slcRefPar, nil, nil, 1)
    
    if err != nil {
        return ferr.Wrap(err)
    }

    dat1, dat2 := slc1.Dat, slc2.Dat
    _, err = slcDiffIntf(dat1, dat2, par1, par2, out.Par,
        out.SimUnwrap, out.DiffPar, rng, azi, 0, 0)
    
    if err != nil {
        return ferr.Wrap(err)
    }
    
    if err = out.Parse(); err != nil {
        return ferr.Wrap(err)
    }
    
    // TODO: Check date difference order
    out.DeltaT = slc1.Time.Sub(slc2.Time)
    
    return nil
}

type CpxToReal int

const (
    Real CpxToReal = iota
    Imaginary
    Intensity
    Magnitude
    Phase
)

func (c CpxToReal) String() string {
    switch c {
    case Real:
        return "Real"
    case Imaginary:
        return "Imaginary"
    case Intensity:
        return "Intensity"
    case Magnitude:
        return "Magnitude"
    case Phase:
        return "Phase"
    default:
        return "Unknown"
    }
}

var cpxToReal = Gamma.Must("cpx_to_real")

func (ifg File) ToReal(mode CpxToReal, name string) (d datafile.File, err error) {
    var ferr = merr.Make("IFG.ToReal")
    
    if len(name) == 0 {
        d, err = TmpDatFile("real", Float)
    } else {
        d, err = NewDatFile(name, Float)
    }
    
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    d.Ra = ifg.Ra
    
    Mode := 0
    
    switch (mode) {
    case Real:
        Mode = 0
    case Imaginary:
        Mode = 1
    case Intensity:
        Mode = 2
    case Magnitude:
        Mode = 3
    case Phase:
        Mode = 4
    default:
        err = ferr.Wrap(ModeError{name:"IFG.ToReal", got:mode})
        return
    }
    
    if _, err = cpxToReal(ifg.Dat, d.Dat, d.Rng, Mode); err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

var rasmph_pwr24 = Gamma.Must("rasmph_pwr24")

func (ifg File) Raster(opt plot.RasArgs) (err error) {
    var ferr = merr.Make("IFG.Raster")

    opt.Mode = MagPhasePwr
    
    if err = ifg.Raster(opt); err != nil {
        err = ferr.Wrap(err)
    }
    return
}

func (ifg File) rng() (i int, err error) {
    var ferr = merr.Make("IFG.rng")
    
    if i, err = ifg.Int("interferogram_width", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return 
}

func (ifg File) azi() (i int, err error) {
    var ferr = merr.Make("IFG.azi")
    
    if i, err = ifg.Int("interferogram_azimuth_lines", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return 
}
