package interferogram

import (
    "time"
    "log"
    "strings"
    "strconv"
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/cli/stream"

    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/base"
    "github.com/bozso/gamma/plot"
)

type File struct {
    data.ComplexFile
    DiffPar   string        `json:"diff_par"`
    Quality   string        `json:"quality"`
    SimUnwrap string        `json:"simulated_unwrap"`
    DeltaT    time.Duration `json:"delta_time"`
}

func New(dat, off, diffpar string) (ifg File) {
    ifg.File.DatFile, ifg.File.ParFile = dat, off

    if len(diffpar) == 0 {
        diffpar = dat + ".diff_par"
    }
    
    ifg.DiffPar = diffpar
    
    return
}

// TODO: implement
var Importer = data.ParamKeys{
    
}

// TODO: implement
func (f *File) Load() (err error) {
    return nil
}

func (ifg File) WithShape(datafile, off, diffpar string) (out File) {
    if len(off) == 0 {
        diffpar = datafile + ".off"
    }

    if len(diffpar) == 0 {
        diffpar = datafile + ".diff_par"
    }
    
    out.File.DatFile, out.File.ParFile, out.DiffPar = datafile, off, diffpar
    
    return
}

func (i File) Move(dir string) (im File, err error) {
    if im.File, err = i.File.Move(dir); err != nil {
        return
    }
    
    if im.DiffPar, err = path.Move(i.DiffPar, dir); err != nil {
        return
    }
    
    if len(i.SimUnwrap) > 0 {
        if im.SimUnwrap, err = path.Move(i.SimUnwrap, dir); err != nil {
            return
        }
    }
    
    if len(i.Quality) > 0 {
        if im.Quality, err = path.Move(i.Quality, dir); err != nil {
            return
        }
    }
    
    return
}


func (ifg File) CheckQuality() (b bool, err error) {
    qual := ifg.Quality
    
    file := stream.In{}
    if err = file.Set(qual); err != nil {
        return
    }
    defer file.Close()
    
    offs := 0.0
    var diff float64

    scan := file.Scanner()
    for scan.Scan() {
        line := scan.Text()
        
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
        Looks common.RngAzi
        interact bool
        hgt, datapath, off, diff string
        algo OffsetAlgo
        ref *base.SLC
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

func FromSLC(slc1, slc2 base.SLC, opt IfgOpt) (out File, err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    par1, par2 := slc1.ParFile, slc2.ParFile
    
    out = New(opt.datapath, opt.off, opt.diff)
    
    // TODO: check arguments!
    _, err = createOffset.Call(par1, par2, out.ParFile, opt.algo,
        rng, azi, inter)
    if err != nil { return }
    
    slcRefPar := "-"
    
    if ref := opt.ref; ref != nil {
        slcRefPar = ref.ParFile
    }
    
    _, err = phaseSimOrb.Call(par1, par2, out.ParFile, opt.hgt,
        out.SimUnwrap, slcRefPar, nil, nil, 1)
    if err != nil { return }

    dat1, dat2 := slc1.DatFile, slc2.DatFile
    _, err = slcDiffIntf.Call(dat1, dat2, par1, par2, out.ParFile,
        out.SimUnwrap, out.DiffPar, rng, azi, 0, 0)
    if err != nil { return }
    
    if err = out.Load(); err != nil {
        return
    }
    
    // TODO: Check date difference order
    out.DeltaT = slc1.Time.Sub(slc2.Time)
    
    return
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

func (ifg File) ToReal(mode CpxToReal, d data.FloatFile) (err error) {
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
        err = utils.UnrecognizedMode(mode.String(), "interferogram.ToReal")
        return
    }
    
    _, err = cpxToReal.Call(ifg.DatFile, d.DataPath(), d.Rng(), Mode)
    
    return
}

var rasmph_pwr24 = common.Must("rasmph_pwr24")

func (ifg File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.MagPhasePwr

    err = ifg.Raster(opt)
    return
}

type AdaptFiltOpt struct {
    offset               common.RngAzi
    alpha, step, frac    float64
    FFTWindow, cohWindow int
}

var adf = common.Must("adf")

func (ifg File) AdaptFilt(opt AdaptFiltOpt, Ifg File, cc Coherence) (err error) {
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
