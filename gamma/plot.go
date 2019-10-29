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
    
    DisArgs struct {
        ScaleExp
        RngAzi
        Minmax
        Flip     bool    `name:"flip" default:""`
        DType
        //Dtype    string  `name:"dtype" default:""`
        Datfile  string  `name:"dat" default:""`
        Cmd      string  `name:"cmd" default:""`
        Start    int     `name:"start" default:"0"`
        Nlines   int     `name:"nlines" default:"1"`
        Sec      string  `name:"sec" default:""`
        StartCC  int     `name:"startcc"  default:"1"`
        StartPwr int     `name:"startpwr" default:"1"`
        StartCpx int     `name:"startcpx" default:"1"`
        StartHgt int     `name:"starthgt" default:"1"`
        Coh      string  `name:"coh" default:""`
        Cycle    float64 `name:"cycle" default:"160.0"`
        LR       int
        Elev     float64 `name:"elev" default:""`
        Orient   float64 `name:"orient" default:""`
        ColPost  float64 `name:"colpost" default:""`
        RowPost  float64 `name:"rowpost" default:""`
        zeroFlag ZeroFlag
    }
)

const (
    Missing ZeroFlag = iota
    Valid
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
    
    if arg.StartPwr == 0 {
        arg.StartPwr = 1
    }
    
    if arg.StartCpx == 0 {
        arg.StartCpx = 1
    }
    
    if arg.StartHgt == 0 {
        arg.StartHgt = 1
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


var _rasslc = Gamma.Must("rasSLC")

func rasslc(opt RasArgs) error {
    dtype := 0
    
    switch opt.DType {
    case FloatCpx:
        dtype = 0
    case ShortCpx:
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s' for rasslc",
            opt.DType.ToString())
    }
    
    _, err := _rasslc(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                      dtype, opt.HeaderSize, opt.Raster)
    
    return err
}

var _raspwr = Gamma.Must("raspwr")

func raspwr(opt RasArgs) error {
    dtype := 0
    
    switch opt.DType {
    case Float:
        dtype = 0
    case Short:
        dtype = 1
    case Double:
        dtype = 2
    default:
        return Handle(nil, "unrecognized image format '%s' for raspwr",
            opt.DType.ToString())
    }
    
    _, err := _raspwr(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                      opt.LR, opt.Raster, dtype, opt.HeaderSize)

    return err
}

var _rasmph = Gamma.Must("rasmph")

func rasmph(opt RasArgs) error {
    dtype := 0
    
    switch opt.DType {
    case FloatCpx:
        dtype = 0
    case ShortCpx:
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s' for rasmph",
            opt.DType.ToString())
    }
    
    _, err := _rasmph(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                      opt.LR, opt.Raster, dtype) 
    
    return err
}


/*
 * TODO: finish implementation
 * 
var _rasshd = Gamma.must("rasshd")

func (opt *shdArgs) Parse(d DataFile) error {
    err := shdArgs.rasArgs.Parse(d)
    
    if err != nil {
        return err
    }
}

func rasshd(opt *shdArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FLOAT":
        dtype = 0
    case "SHORT INTEGER":
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s'", opt.ImgFmt)
    }
    
    _, err := _rasshd(opt.Datfile, opt.Rng, colPost, RowPost, opt.Start,
                      opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.Elev,
                      opt.Orient opt.LR, opt.raster, dtype, opt.zeroFlag) 
    
    return err
}
*/
