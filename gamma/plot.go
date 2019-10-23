package gamma

import (
    "fmt"
    //"log"
)

type(
    ScaleExp struct {
        Scale float64 `name:"scale" default:"1.0"`
        Exp   float64 `name:"exp" default:"0.35"`
    }
    
    DisArgs struct {
        ScaleExp
        RngAzi
        Minmax
        Flip     bool    `name:"flip" default:""`
        ImgFmt   string  `name:"imgfmt" default:""`
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
    }

    RasArgs struct {
        DisArgs
        AvgFact    int    `name:"afact" default:"1000"`
        HeaderSize int    `name:"header" default:"0"`
        Avg        RngAzi `name:"avg"`
        Raster     string `name:"ras"`
    }
    
    ZeroFlag int
    
    ShdArgs struct {
        RasArgs
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

func (arg *DisArgs) Parse(dat DataFile) (err error) {
    arg.ScaleExp.Parse()
    
    if arg.Start == 0 {
        arg.Start = 1
    }
    
    if len(arg.Datfile) == 0 {
        arg.Datfile = dat.Datfile()
    }
    
    if len(arg.Cmd) == 0 {
        arg.Cmd = dat.PlotCmd()
    }

    if arg.Rng == 0 {
        arg.Rng = dat.GetRng()
    }

    if arg.Azi == 0 {
        arg.Azi = dat.GetAzi()
    }

    // parts = pth.basename(datfile).split(".")
    //if len(arg.ImgFmt) == 0 {
        //if arg.ImgFmt, err = dat.ImageFormat(); err != nil {
            //return Handle(err, "failed to get image_format")
        //}
    //}

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

    return nil
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

func (opt *RasArgs) Parse(dat DataFile) error {
    err := opt.DisArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display arguments")
    }
    
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
    
    return nil
}


var _rasslc = Gamma.Must("rasSLC")

func rasslc(opt RasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FCOMPLEX":
        dtype = 0
    case "SCOMPLEX":
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s' for rasslc",
            opt.ImgFmt)
    }
    
    _, err := _rasslc(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                      dtype, opt.HeaderSize, opt.Raster)
    
    return err
}

var _raspwr = Gamma.Must("raspwr")

func raspwr(opt RasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FLOAT":
        dtype = 0
    case "SHORT INTEGER":
        dtype = 1
    case "DOUBLE":
        dtype = 2
    default:
        return Handle(nil, "unrecognized image format '%s' for raspwr",
            opt.ImgFmt)
    }
    
    _, err := _raspwr(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                      opt.LR, opt.Raster, dtype, opt.HeaderSize)

    return err
}

var _rasmph = Gamma.Must("rasmph")

func rasmph(opt RasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FCOMPLEX":
        dtype = 0
    case "SCOMPLEX":
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s' for rasmph",
            opt.ImgFmt)
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
    
    if op.Elev == 0.0 {
        op.Elev = 45.0
    }
    
    if op.Orient == 0.0 {
        op.Orient = 135.0
    }
    
    if op.colPost == 0.0 {
        op.colPost, err = d.Float()
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


func (se *ScaleExp) Parse() {
    if se.Scale == 0.0 {
        se.Scale = 1.0
    }
    
    if se.Exp == 0.0 {
        se.Exp = 0.35
    }
}
