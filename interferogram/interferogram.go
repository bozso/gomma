package interferogram

import (
    "time"
    "log"
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/splitted"

    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/geo"
    "github.com/bozso/gomma/plot"
)

type File struct {
    data.ComplexWithPar       `json:"complex_file"`
    DiffPar   path.ValidFile  `json:"diff_par"`
    Quality   path.File       `json:"quality"`
    SimUnwrap path.File       `json:"simulated_unwrap"`
    DeltaT    time.Duration   `json:"delta_time"`
}

func (i File) Move(dir path.Dir) (im File, err error) {
    if im.File, err = i.File.Move(dir); err != nil {
        return
    }
    
    if im.DiffPar, err = i.DiffPar.Move(dir); err != nil {
        return
    }
    
    f, err := i.SimUnwrap.ToValid()
    
    if err != nil {
        f, err = f.Move(dir)
        
        if err != nil {
            return
        }
        
        im.SimUnwrap = f.ToFile()
    }

    f, err = i.Quality.ToValid()
    
    if err != nil {
        f, err = f.Move(dir)
        
        if err != nil {
            return
        }
        
        im.Quality = f.ToFile()
    }
    
    im.Meta, im.DeltaT = i.Meta, im.DeltaT    
    return
}

func (ifg File) CheckQuality() (b bool, err error) {
    qual, err := ifg.Quality.ToValid()
    if err != nil {
        return
    }
    
    scan, err := qual.Scanner()
    if err != nil {
        return
    }
    defer scan.Close()
    
    offs := 0.0

    for scan.Scan() {
        line := scan.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split, Err := splitted.NewFields(line)
        if err != nil {
            err = Err
            return
        }
        
        first, Err := split.Idx(0)
        if err != nil {
            err = Err
            return
        }
        
        if first == "azimuth_pixel_offset" {
            diff, Err := split.Float(1)
            if err != nil {
                err = Err
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

type OffsetAlgo int

type IfgOpt struct {
    Looks common.RngAzi
    interact bool
    hgt geo.Height
    datapath, off, diff path.File
    algo OffsetAlgo
    ref *slc.SLC
}

const (
    IntensityCoherence OffsetAlgo = iota
    FingeVisibility
)

var (
    createOffset  = common.Must("create_offset")
    phaseSimOrb   = common.Must("phase_sim_orb")
    slcDiffIntf   = common.Must("SLC_diff_intf")
)

func FromSLC(slc1, slc2 slc.SLC, opt IfgOpt) (out File, err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    par1, par2, ra := slc1.ParFile, slc2.ParFile, opt.Looks
    
    p := New(opt.datapath.Path)
    
    // TODO: check arguments!
    _, err = createOffset.Call(par1, par2, p.ParFile, opt.algo,
        ra.Rng, ra.Azi, inter)
    
    if err != nil {
        return
    }
    
    slcRefPar := "-"
    
    if ref := opt.ref; ref != nil {
        slcRefPar = ref.ParFile.GetPath()
    }
    
    _, err = phaseSimOrb.Call(par1, par2, out.ParFile, opt.hgt,
        out.SimUnwrap, slcRefPar, nil, nil, 1)
    if err != nil { return }

    dat1, dat2 := slc1.DatFile, slc2.DatFile
    
    _, err = slcDiffIntf.Call(dat1, dat2, par1, par2, out.ParFile,
        out.SimUnwrap, out.DiffPar, ra.Rng, ra.Azi, 0, 0)
    
    if err != nil {
        return
    }
    
    out, err = p.Load()
    if err != nil {
        return
    }
    
    // TODO: Check date difference order
    out.DeltaT = slc1.Time.Sub(slc2.Time)
    
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
