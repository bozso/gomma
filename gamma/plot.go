package gamma

type(
    ScaleExp struct {
        Scale, Exp float64
    }
    
    disArgs struct {
        Flip                 bool
        ImgFmt, Datfile, Cmd string
        Start, Nlines, LR    int
        ScaleExp
        RngAzi
    }

    rasArgs struct {
        disArgs
        avgFact, headerSize int
        Avg                 RngAzi
        raster              string
    }
)

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
