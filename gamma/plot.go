package gamma

import (
    "fmt"
    //"log"
)

type(
    ScaleExp struct {
        Scale, Exp float64
    }
    
    disArgs struct {
        ScaleExp
        RngAzi
        Flip                 bool
        ImgFmt, Datfile, Cmd string
        Start, Nlines, LR    int
    }

    rasArgs struct {
        disArgs
        avgFact, headerSize int
        Avg                 RngAzi
        raster              string
    }
    
    ZeroFlag int
    
    shdArgs struct {
        rasArgs
        Elev, Orient, colPost, rowPost float64
        zeroFlag ZeroFlag
    }
)

const (
    Missing ZeroFlag = iota
    Valid
)

func (arg *disArgs) Parse(dat DataFile) (err error) {
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
        if arg.Rng, err = dat.Rng(); err != nil {
            return Handle(err, "failed to get range_samples")
        }
    }

    if arg.Azi == 0 {
        if arg.Azi, err = dat.Azi(); err != nil {
            return Handle(err, "failed to get azimuth_lines")
        }
    }

    // parts = pth.basename(datfile).split(".")
    if len(arg.ImgFmt) == 0 {
        if arg.ImgFmt, err = dat.ImageFormat(); err != nil {
            return Handle(err, "failed to get image_format")
        }
    }

    if arg.Flip {
        arg.LR = -1
    } else {
        arg.LR = 1
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

func (opt *rasArgs) Parse(dat DataFile) error {
    err := opt.disArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display arguments")
    }
    
    if opt.avgFact == 0 {
        opt.avgFact = 1000
    }
    
    if opt.Avg.Rng == 0 {
        opt.Avg.Rng = calcFactor(opt.Rng, opt.avgFact)
    }
    
    if opt.Avg.Azi == 0 {
        opt.Avg.Azi = calcFactor(opt.Azi, opt.avgFact)
    }
    
    if len(opt.raster) == 0 {
        opt.raster = fmt.Sprintf("%s.%s", opt.Datfile, Settings.RasExt)
    }
    
    return nil
}


var _rasslc = Gamma.must("rasSLC")

func rasslc(opt rasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FCOMPLEX":
        dtype = 0
    case "SCOMPLEX":
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s'", opt.ImgFmt)
    }
    
    _, err := _rasslc(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                      dtype, opt.headerSize, opt.raster)
    
    return err
}

var _raspwr = Gamma.must("raspwr")

func raspwr(opt rasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FLOAT":
        dtype = 0
    case "SHORT INTEGER":
        dtype = 1
    case "DOUBLE":
        dtype = 2
    default:
        return Handle(nil, "unrecognized image format '%s'", opt.ImgFmt)
    }
    
    _, err := _raspwr(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                      opt.LR, opt.raster, dtype, opt.headerSize)

    return err
}

var _rasmph = Gamma.must("rasmph")

func rasmph(opt rasArgs) error {
    dtype := 0
    
    switch opt.ImgFmt {
    case "FCOMPLEX":
        dtype = 0
    case "SCOMPLEX":
        dtype = 1
    default:
        return Handle(nil, "unrecognized image format '%s'", opt.ImgFmt)
    }
    
    _, err := _rasmph(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                      opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                      opt.LR, opt.raster, dtype) 
    
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

func rasshd(opt shdArgs) error {
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
