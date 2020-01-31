package plot

import (
    "fmt"
    
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/data"
)

type RasArgs struct {
    DisArgs
    AvgFact    int    `name:"afact" default:"1000"`
    HeaderSize int    `name:"header" default:"0"`
    Avg        common.RngAzi `name:"avg"`
    Raster     string `name:"ras"`
}

func (opt *RasArgs) Parse(dat data.IFile) {
    opt.DisArgs.Parse(dat)
    
    if opt.AvgFact == 0 {
        opt.AvgFact = 1000
    }
    
    if opt.Avg.Rng == 0 {
        opt.Avg.Rng = calcFactor(opt.Rng, opt.AvgFact)
    }
    
    if opt.Avg.Azi == 0 {
        opt.Avg.Azi = calcFactor(opt.Azi, opt.AvgFact)
    }
    
    if len(opt.Raster) == 0 {
        opt.Raster = fmt.Sprintf("%s.%s", opt.Datfile,
            common.Settings.RasExt)
    }    
}

type PlotMode int

const (
    Byte PlotMode = iota
    CC
    Decibel
    Deform
    Height
    Linear
    MagPhase
    MagPhasePwr
    Power
    SingleLook
    Unwrapped
    Undefined
)

func Raster(d data.IFile, opt RasArgs) (err error) {
    opt.Parse(d)
    
    //fmt.Printf("%#v\n", opt)
    //return nil
                
    switch opt.Mode {
    case Byte:
        _, err = rasByte(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                         opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.LR,
                         opt.Raster)
    case CC:
        _, err = rasCC(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                       opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                       opt.Min, opt.Max, opt.Scale, opt.Exp, opt.LR,
                       opt.Raster)
    //case Decibel:
        //_, err = rasdB(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                       //opt.Avg.Rng, opt.Avg.Azi, opt.Min, opt.Max,
                       //opt.Offset, opt.LR, opt.Raster, opt.AbsFlag,
                       //opt.Inverse, opt.Channel)
    case Deform:
        _, err = rasdtPwr(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                          opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                          opt.Cycle, opt.Scale, opt.Exp, opt.LR, opt.Raster,
                          opt.CC, opt.StartCC, opt.CCMin)
    case Height:
        _, err = rasHgt(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                        opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                        opt.Cycle, opt.Scale, opt.Exp, opt.LR, opt.Raster)
    case Linear:
        _, err = rasLinear(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                           opt.Avg.Rng, opt.Avg.Azi, opt.Min, opt.Max, opt.LR,
                           opt.Raster, opt.Inverse, opt.Channel)
    case MagPhase:
        dt := 0
        
        switch opt.Type {
        case data.FloatCpx:
            dt = 0
        case data.ShortCpx:
            dt = 1
        default:
            // Error
        }
        _, err = rasMph(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                        opt.LR, opt.Raster, dt)
    case MagPhasePwr:    
        if opt.Type != data.FloatCpx {
            // Error
        }
        
        _, err = rasMphPwr(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                           opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                           opt.Scale, opt.Exp, opt.LR, opt.Raster,
                           opt.CC, opt.StartCC, opt.CCMin)
    case Power:
        dt := 0
        
        switch opt.Type {
        case data.Float:
            dt = 0
        case data.Short:
            dt = 1
        case data.Double:
            dt = 2
        default:
            // Error
        }
        
        _, err = rasPwr(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                        opt.Raster, dt, opt.HeaderSize)
    
    case SingleLook:
        dt := 0
        
        switch opt.Type {
        case data.FloatCpx:
            dt = 0
        case data.ShortCpx:
            dt = 1
        default:
            // Error
        }
        
        _, err = rasSLC(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                        opt.LR, dt, opt.HeaderSize, opt.Raster)
    case Unwrapped:
        _, err = rasRmg(opt.Datfile, opt.Sec, opt.Start, opt.StartSec, 
                        opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.PhaseScale,
                        opt.Scale, opt.Exp, opt.Offset, opt.LR, opt.Raster,
                        opt.CC, opt.StartCC, opt.CCMin)
    }
    return err
}

