package interferogram

import (
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/datafile"
    "github.com/bozso/gamma/plot"
)

type File struct {
    data.File
    DiffPar   string        `json:"diff_par"`
    Quality   string        `json:"quality"`
    SimUnwrap string        `json:"simulated_unwrap"`
    DeltaT    time.Duration `json:"delta_time"`
}

func New(dat, off, diffpar string) (ifg IFG) {
    ifg.File.DatFile, ifg.File.ParFile = dat, off

    if len(diffpar) == 0 {
        diffpar = dat + ".diff_par"
    }
    
    ifg.DiffPar = diffpar
    
    return
}

func (i File) Move(dir string) (im File, err error) {
    if im.DatParFile, err = i.DatParFile.Move(dir); err != nil {
        return
    }
    
    if im.DiffPar.Par, err = Move(i.DiffPar.Par, dir); err != nil {
        return
    }
    
    if len(i.SimUnwrap) > 0 {
        if im.SimUnwrap, err = Move(i.SimUnwrap, dir); err != nil {
            return
        }
    }
    
    if len(i.Quality) > 0 {
        if im.Quality, err = Move(i.Quality, dir); err != nil {
            return
        }
    }
    
    return
}

func (i *File) Validate() (err error) {
    return i.TypeCheck("IFG", "complex", ShortCpx, FloatCpx)
}

func (ifg File) CheckQuality() (b bool, err error) {
    qual := ifg.Quality
    
    file, err := NewReader(qual)
    if err != nil { return }
    
    defer file.Close()
    
    offs := 0.0
    var diff float64

    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := strings.Fields(line)
        
        if split[0] == "azimuth_pixel_offset" {
            s := split[1]
            diff, err = strconv.ParseFloat(s, 64)
            
            if err != nil {
                err = utils.WrapFmt(err,
                    "failed to parse: '%s' into float64", s)
                return
            }
            
            offs += diff
        }
    }
    
    log.Printf("Sum of azimuth offsets in %s is %f pixel.\n", qual, offs)
    
    if offs > 0.0 || offs < 0.0 {
        b = true
    } else {
        b = false
    }
    
    return
}

type ( 
    OffsetAlgo int

    IfgOpt struct {
        Looks RngAzi
        interact bool
        hgt string
        algo OffsetAlgo
    }
)

const (
    IntensityCoherence OffsetAlgo = iota
    FingeVisibility
)

var (
    createOffset  = common.Must("create_offset")
    phaseSimOrb   = common.Must("phase_sim_orb")
    slcDiffIntf   = common.Must("SLC_diff_intf")
)

func FromSLC(slc1, slc2, ref *base.SLC, out File, opt IfgOpt) (err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    par1, par2 := slc1.Par, slc2.Par
    
    // TODO: check arguments!
    _, err = createOffset.Call(par1, par2, out.Par, opt.algo,
        rng, azi, inter)
    if err != nil { return }
    
    slcRefPar := "-"
    
    if ref != nil {
        slcRefPar = ref.Par
    }
    
    _, err = phaseSimOrb.Call(par1, par2, out.Par, opt.hgt,
        out.SimUnwrap, slcRefPar, nil, nil, 1)
    if err != nil { return }

    dat1, dat2 := slc1.Dat, slc2.Dat
    _, err = slcDiffIntf(dat1, dat2, par1, par2, out.Par,
        out.SimUnwrap, out.DiffPar, rng, azi, 0, 0)
    if err != nil { return }
    
    if err = out.Parse(); err != nil {
        return
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

var cpxToReal = common.Must("cpx_to_real")

func (ifg File) ToReal(mode CpxToReal, d data.Data) (err error) {
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
        err = ModeError{name:"IFG.ToReal", got:mode}
        return
    }
    
    _, err = cpxToReal.Call(ifg.DatFile, d.DataPath(), d.Rng(), Mode)
    
    return
}

var rasmph_pwr24 = common.Must("rasmph_pwr24")

func (ifg File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = MagPhasePwr

    err = ifg.Raster(opt)
    return
}

type AdaptFiltOpt struct {
    offset               common.RngAzi
    alpha, step, frac    float64
    FFTWindow, cohWindow int
}

var adf = common.Must("adf")

func (ifg IFG) AdaptFilt(opt AdaptFiltOpt, Ifg IFG, cc Coherence) (err error) {
    step := float64(opt.FFTWindow) / 8.0
    
    if opt.step > 0.0 {
        step = opt.step
    }
    
    _, err = adf.Call(
        ifg.DatFile, Ifg.DatFile, cc.DatFile, ifg.Ra.Rng,
        opt.alpha, opt.FFTWindow, opt.cohWindow, step,
        opt.offset.Azi, opt.offset.Rng, opt.frac)
    return
}
