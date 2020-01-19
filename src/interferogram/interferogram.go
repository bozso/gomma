package interferogram

import (
    "../common"
    "../datafile"
    "../plot"
)

type File struct {
    data.File
    DiffPar   string
    Quality   string
    SimUnwrap string
    DeltaT    time.Duration
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

func (ifg File) CheckQuality() (b bool, err error) {
    var (
        ferr = merr.Make("IFG.CheckQuality")
        qual = ifg.Quality
    )
    
    var file Reader
    if file, err = NewReader(qual); err != nil {
        err = ferr.Wrap(err)
        return
    }
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
                err = ferr.WrapFmt(err,
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
    createOffset  = Gamma.Must("create_offset")
    phaseSimOrb   = Gamma.Must("phase_sim_orb")
    slcDiffIntf   = Gamma.Must("SLC_diff_intf")
)

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

type AdaptFiltOpt struct {
    offset               common.RngAzi
    alpha, step, frac    float64
    FFTWindow, cohWindow int
}

var adf = Gamma.Must("adf")

//func (ifg IFG) AdaptFilt(opt AdaptFiltOpt) (Ifg IFG, cc Coherence, err error) {
    //step := float64(opt.FFTWindow) / 8.0
    
    //if opt.step > 0.0 {
        //step = opt.step
    //}
    
    //// TODO: figure out the name of the output files
    //ret, err = NewIFG(self.Dat + ".filt", "", "", "", "")
    
    //if err != nil {
        //err = Handle(err, "failed to create new interferogram struct")
        //return
    //}
    
    //cc, err = NewCoherence("", "")
    
    //if err != nil {
        //err = Handle(err, "failed to create new dataFile struct")
        //return
    //}
    
    ///*
    //if Empty(filt):
        //filt = 
    
    //if empty(cc is None:
        //cc = self.datfile + ".cc"
    //*/
    
    //rng := self.Rng
    
    //_, err = adf(self.Dat, ret.Dat, cc.Dat, rng, opt.alpha, opt.FFTWindow,
                 //opt.cohWindow, step, opt.offset.Azi, opt.offset.Rng,
                 //opt.frac)
    
    //if err != nil {
        //err = Handle(err, "adaptive filtering failed")
        //return
    //}
    
    //return ret, cc, nil
//}
