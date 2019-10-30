package gamma

import (
    "fmt"
    //"log"
)

type ScaleExp struct {
    Scale float64 `name:"scale" default:"1.0"`
    Exp   float64 `name:"exp" default:"0.35"`
}

func (se *ScaleExp) Parse() {
    if se.Scale == 0.0 {
        se.Scale = 1.0
    }
    
    if se.Exp == 0.0 {
        se.Exp = 0.35
    }
}

type (
    ZeroFlag int
    Channel int
    Inverse int
    
    DisArgs struct {
        ScaleExp
        RngAzi
        Minmax
        DType
        Inverse
        Channel
        Mode       PlotMode
        zeroFlag   ZeroFlag
        Flip       bool    `name:"flip" default:""`
        Datfile    string  `name:"dat" default:""`
        Start      int     `name:"start" default:"0"`
        Nlines     int     `name:"nlines" default:"0"`
        Sec        string  `name:"sec" default:""`
        StartSec   int     `name:"startSec" default:"1"`
        StartCC    int     `name:"startCC" default:"1"`
        Coh        string  `name:"coh" default:""`
        Cycle      float64 `name:"cycle" default:"160.0"`
        LR         int
        Elev       float64 `name:"elev" default:"45.0"`
        Orient     float64 `name:"orient" default:"135.0"`
        ColPost    float64 `name:"colpost" default:"0"`
        RowPost    float64 `name:"rowpost" default:"0"`
        Offset     float64 `name:"offset" default:"0.0"`
        PhaseScale float64 `name:"scale" default:"0.0"`
        CC         string
        CCMin      float64 `name:"ccMin" default:"0.2"`
    }
)

const (
    Missing ZeroFlag = iota
    Valid
)

const (
    Float2Raster Inverse = 1
    Raster2Float Inverse = -1
)

const (
    Red   Channel = 1
    Green Channel = 2
    Blue  Channel = 3
)

func (arg *DisArgs) Parse(dat IDatFile) {
    arg.ScaleExp.Parse()
    
    if arg.Start == 0 {
        arg.Start = 1
    }
    
    if len(arg.Datfile) == 0 {
        arg.Datfile = dat.Datfile()
    }
    
    if arg.Rng == 0 {
        arg.Rng = dat.Rng()
    }

    if arg.Azi == 0 {
        arg.Azi = dat.Azi()
    }
    
    arg.DType = dat.Dtype()
        
    if arg.Flip {
        arg.LR = -1
    } else {
        arg.LR = 1
    }
    
    if arg.Min == 0.0 {
        arg.Min = 0.1
    }
    
    if arg.Max == 0.0 {
        arg.Min = 0.9
    }
    
    if arg.StartCC == 0 {
        arg.StartCC = 1
    }
    
    if arg.StartSec == 0 {
        arg.StartSec = 1
    }
    
    
    if arg.Cycle == 0 {
        arg.Cycle = 160.0
    }
    
    if arg.Elev == 0.0 {
        arg.Elev = 45.0
    }
    
    if arg.Orient == 0.0 {
        arg.Orient = 135.0
    }
    
    //if arg.Mode == Undefined {
        //switch opt.DType {
            
        //}
    //}
    
    if arg.Inverse == 0 {
        arg.Inverse = Float2Raster 
    }
    
    if arg.Channel == 0 {
        arg.Channel = Red
    }
    
    if arg.CCMin == 0.0 {
        arg.CCMin = 0.2
    }
    
    //if op.colPost == 0.0 {
        //op.colPost, err = d.Float()
    //}
}

func calcFactor(ndata, factor int) int {
    // log.Printf("ndata: %d factor: %d\n", ndata, factor)
    
    ret := float64(ndata) / float64(factor)
    
    // log.Fatalf("ret: %f\n", ret)
    
    if ret <= 0.0 {
        return 1
    } else {
        return int(ret)
    }
}

type RasArgs struct {
    DisArgs
    AvgFact    int    `name:"afact" default:"1000"`
    HeaderSize int    `name:"header" default:"0"`
    Avg        RngAzi `name:"avg"`
    Raster     string `name:"ras"`
}

func (opt *RasArgs) Parse(dat IDatFile) {
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
        opt.Raster = fmt.Sprintf("%s.%s", opt.Datfile, Settings.RasExt)
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

func (d DatFile) Raster(opt RasArgs) (err error) {
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
        
        switch opt.DType {
        case FloatCpx:
            dt = 0
        case ShortCpx:
            dt = 1
        default:
            // Error
        }
        _, err = rasMph(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                        opt.LR, opt.Raster, dt)
    case MagPhasePwr:    
        if opt.DType != FloatCpx {
            // Error
        }
        
        _, err = rasMphPwr(opt.Datfile, opt.Sec, opt.Rng, opt.Start,
                           opt.StartSec, opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                           opt.Scale, opt.Exp, opt.LR, opt.Raster,
                           opt.CC, opt.StartCC, opt.CCMin)
    case Power:
        dt := 0
        
        switch opt.DType {
        case Float:
            dt = 0
        case Short:
            dt = 1
        case Double:
            dt = 2
        default:
            // Error
        }
        
        _, err = rasPwr(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                        opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                        opt.Raster, dt, opt.HeaderSize)
    
    case SingleLook:
        dt := 0
        
        switch opt.DType {
        case FloatCpx:
            dt = 0
        case ShortCpx:
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


var (
    rasByte = Gamma.Must("rasbyte")
    rasCC = Gamma.Must("rascc")
    rasdB = Gamma.Must("ras_dB")
    rasHgt = Gamma.Must("rashgt")
    rasdtPwr = Gamma.Must("rasdt_pwr")
    rasMph = Gamma.Must("rasmph")
    rasMphPwr = Gamma.Must("rasmph_pwr")
    rasPwr = Gamma.Must("raspwr")
    rasRmg = Gamma.Must("rasrmg")
    rasShd = Gamma.Must("rasshd")
    rasSLC = Gamma.Must("rasSLC")
    rasLinear = Gamma.Must("ras_linear")
)
