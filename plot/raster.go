package plot

import (
    "github.com/bozso/gotoolbox/command"

    "github.com/bozso/gomma/data"
)

func Raster(cmd command.Command, p Plottable, opt RasArgs) (err error) {
    opt.Parse(p)
    
    //fmt.Printf("%#v\n", opt)
    //return nil
                
    switch opt.Mode {
    case Byte:
        _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                         opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.LR,
                         opt.Raster)
    case CC:
        _, err = cmd.Call(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                       opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                       opt.Min, opt.Max, opt.Scale, opt.Exp, opt.LR,
                       opt.Raster)
    //case Decibel:
        //_, err = rasdB.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                       //opt.Avg.Rng, opt.Avg.Azi, opt.Min, opt.Max,
                       //opt.Offset, opt.LR, opt.Raster, opt.AbsFlag,
                       //opt.Inverse, opt.Channel)
    case Deform:
        _, err = cmd.Call(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                          opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                          opt.Cycle, opt.Scale, opt.Exp, opt.LR, opt.Raster,
                          opt.CC, opt.StartCC, opt.CCMin)
    case Height:
        _, err = cmd.Call(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                        opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                        opt.Cycle, opt.Scale, opt.Exp, opt.LR, opt.Raster)
    case Linear:
        _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
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
        _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                        opt.LR, opt.Raster, dt)
    case MagPhasePwr:    
        if opt.Type != data.FloatCpx {
            // Error
        }
        
        _, err = cmd.Call(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
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
        
        _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
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
        
        _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                        opt.LR, dt, opt.HeaderSize, opt.Raster)
    case Unwrapped:
        _, err = cmd.Call(opt.Datfile, opt.Sec, opt.Start, opt.StartSec, 
                        opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.PhaseScale,
                        opt.Scale, opt.Exp, opt.Offset, opt.LR, opt.Raster,
                        opt.CC, opt.StartCC, opt.CCMin)
    }
    return err
}
