package dem

import (
    "fmt"
    
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/plot"
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/params"
)

type Lookup struct {
    data.ComplexFile
}

var coord2sarpix = common.Must("coord_to_sarpix")

func ToRadar(ll common.LatLon, mpar, hgt, diffPar string) (ra common.RngAzi, err error) {
    const par = "corrected SLC/MLI range, azimuth pixel (int)"
    
    out, err := coord2sarpix.Call(mpar, "-", ll.Lat, ll.Lon, hgt, diffPar)
    if err != nil {
        err = utils.WrapFmt(err, "failed to retreive radar coordinates")
        return
    }
    
    param := params.FromString(out, ":")
    
    line, err := param.Param(par)
    
    if err != nil {
        err = utils.WrapFmt(err, "failed to retreive range, azimuth")
        return
    }
    
    
    split, err := utils.NewSplitParser(line, " ")
    if err != nil { return }
    
    if split.Len() < 2 {
        err = fmt.Errorf("split to retreive range, azimuth failed")
        return
    }
    
    ra.Rng, err = split.Int(0)
    if err != nil { return }
    
    ra.Azi, err = split.Int(1)
    if err != nil { return }
    
    return ra, nil
}

func (l Lookup) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.MagPhase
    opt.Parse(l)

    err = plot.Raster(l, opt)
    return
}

type InterpolationMode int

const (
    NearestNeighbour InterpolationMode = iota
    BicubicSpline
    BicubicSplineLog
    BicubicSplineSqrt
    BSpline
    BSplineSqrt
    Lanczos
    LanczosSqrt
    InvDist
    InvSquaredDist
    Constant
    Gauss
)

func (i *InterpolationMode) Decode(s string) (err error) {
    switch s {
    case "NearestNeighbour":
        *i = NearestNeighbour
    case "BicubicSpline":
        *i = BicubicSpline
    case "BicubicSplineLog":
        *i = BicubicSplineLog
    case "BicubicSplineLogSqrt":
        *i = BicubicSplineSqrt
    case "BSpline":
        *i = BSpline
    case "BSplineSqrt":
        *i = BSplineSqrt
    case "Lanczos":
        *i = Lanczos
    case "LanczosSqrt":
        *i = LanczosSqrt
    case "InverseDistance":
        *i = InvDist
    case "InverseSquaredDistance":
        *i = InvSquaredDist
    case "Constant":
        *i = Constant
    case "Gauss":
        *i = Gauss
    default:
        err = utils.UnrecognizedMode(s, "Interpolation Mode")
        return
    }
    
    return nil
}

func (i InterpolationMode) String() string {
    switch i {
    case NearestNeighbour:
        return "NearestNeighbour"
    case BicubicSpline:
        return "BicubicSpline"
    case BicubicSplineLog:
        return "BicubicSplineLog"
    case BicubicSplineSqrt:
        return "BicubicSplineLogSqrt"
    case BSpline:
        return "BSpline"
    case BSplineSqrt:
        return "BSplineSqrt"
    case Lanczos:
        return "Lanczos"
    case LanczosSqrt:
        return "LanczosSqrt"
    case InvDist:
        return "InverseDistance"
    case InvSquaredDist:
        return "InverseSquaredDistance"
    case Constant:
        return "Constant"
    case Gauss:
        return "Gauss"
    default:
        return "Unknown"
    }
}

type CodeOpt struct {
    common.RngAzi
    Nlines       int               `cli:"nlines" dft:"0"`
    Npoints      int               `cli:"n,npoint" dft:"4"`
    Oversamp     float64           `cli:"o,oversamp" dft:"2.0"`
    MaxRad       float64           `cli:"m,maxRadious" dft:"0.0"`
    InterpolMode InterpolationMode `cli:"int,interpol dft:"NearestNeighbour"`
    FlipInput    bool              `cli:"flipIn"`
    FlipOutput   bool              `cli:"flipOut"`
    Order        int               `cli:"r,order" dft:"5"`
}

func (co *CodeOpt) SetCli(c *utils.Cli) {
    //c.Var()
    
}

func (opt *CodeOpt) Parse() (lrIn int, lrOut int) {
    lrIn, lrOut = 1, 1
    
    if opt.FlipInput {
        lrIn = -1
    }
    
    if opt.FlipOutput {
        lrOut = -1
    }
    
    if opt.Order == 0 {
        opt.Order = 5
    }
    
    if opt.Oversamp == 0.0 {
        opt.Oversamp = 2.0
    }
    
    if opt.MaxRad == 0.0 {
        opt.MaxRad = 4 * opt.Oversamp
    }
    
    if opt.Npoints == 0 {
        opt.Npoints = 4
    }
        
    return lrIn, lrOut
}


var g2r = common.Must("geocode")

func (l Lookup) geo2radar(in, out data.Data, opt CodeOpt) (err error) {
    lrIn, lrOut := opt.Parse()
    
    if err = opt.RngAzi.Validate(); err != nil {
        return
    }
    
    intm := opt.InterpolMode
    interp := 0

    switch intm {
    case InvDist:
        interp = 0
    case NearestNeighbour:
        interp = 1
    case InvSquaredDist:
        interp = 2
    case Constant:
        interp = 3
    case Gauss:
        interp = 4
    default:
        return utils.UnrecognizedMode(intm.String(), "Interpolation Mode")
    }
    
    dt, dtype := 0, in.DataType()
    
    switch dtype {
    case data.Float:
        dt = 0
    case data.FloatCpx:
        dt = 1
    case data.Raster:
        dt = 2
    case data.UChar:
        dt = 3
    case data.Short:
        dt = 4
    case data.ShortCpx:
        dt = 5
    case data.Double:
        dt = 6
    default:
        return data.WrongType(dtype, "geo2radar")
    }
    
    
    _, err = g2r.Call(l.DatFile, in.DataPath(), in.Rng(),
                 out.DataPath(), out.Rng(),
                 opt.Nlines, interp, dt, lrIn, lrOut, opt.Oversamp,
                 opt.MaxRad, opt.Npoints)
    
    return
}

var r2g = common.Must("geocode_back")

func (l Lookup) radar2geo(in, out data.Data, opt CodeOpt) (err error) {
    lrIn, lrOut := opt.Parse()
    
    if err = opt.RngAzi.Validate(); err != nil {
        return
    }
    
    intm := opt.InterpolMode
    var interp int
    
    // default interpolation mode
    if intm == NearestNeighbour {
        interp = 1
    } else {
        switch intm {
        case NearestNeighbour:
            interp = 0
        case BicubicSpline:
            interp = 1
        case BicubicSplineLog:
            interp = 2
        case BicubicSplineSqrt:
            interp = 3
        case BSpline:
            interp = 4
        case BSplineSqrt:
            interp = 5
        case Lanczos:
            interp = 6
        case LanczosSqrt:
            interp = 7
        default:
            return utils.UnrecognizedMode(intm.String(), "interpolation option")
        }
    }
    
    
    dt, dtype := 0, in.DataType()
    
    switch dtype {
    case data.Float:
        dt = 0
    case data.FloatCpx:
        dt = 1
    case data.Raster:
        dt = 2
    case data.UChar:
        dt = 3
    case data.Short:
        dt = 4
    case data.Double:
        dt = 5
    default:
        return data.WrongType(dtype, "radar2geo")
        
    }
    
    _, err = r2g.Call(in.DataPath(), in.Rng(), l.DatFile,
                 out.DataPath(), out.Rng(),
                 opt.Nlines, interp, dt, lrIn, lrOut, opt.Order)
    
    return
}
