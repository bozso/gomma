package dem

import (
    "../data"
    "../common"
    "../plot"
)

type Lookup struct {
    data.File
}

func (l *Lookup) Set(s string) (err error) {
    return LoadJson(s, l)
}

var coord2sarpix = common.Gamma.Must("coord_to_sarpix")

func (ll LatLon) ToRadar(mpar, hgt, diffPar string) (ra RngAzi, err error) {
    var ferr = merr.Make("LatLon.ToRadar")
    const par = "corrected SLC/MLI range, azimuth pixel (int)"
    
    out, err := coord2sarpix(mpar, "-", ll.Lat, ll.Lon, hgt, diffPar)
    if err != nil {
        err = ferr.WrapFmt(err, "failed to retreive radar coordinates")
        return
    }
    
    params := FromString(out, ":")
    
    line, err := params.Param(par)
    
    if err != nil {
        err = ferr.WrapFmt(err, "failed to retreive range, azimuth")
        return
    }
    
    
    split := NewSplitParser(line, " ")
    
    if err = split.Wrap(); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if len(split.split) < 2 {
        err = ferr.Wrap(fmt.Errorf("split to retreive range, azimuth failed"))
        return
    }
    
    ra.Rng = split.Int(0)
    ra.Azi = split.Int(1)
    
    if err = split.Wrap(); err != nil {
        err = ferr.Wrap(err)
    }
    
    return ra, nil
}

func (l Lookup) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = MagPhase
    opt.Parse(l)

    err = Raster(l, opt)
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
    var ferr = merr.Make("InterpolationMode.Decode")
    
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
        err = ferr.Wrap(UnrecognizedMode{got: s, name:"Interpolation Mode"})
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
    RngAzi
    Nlines       int               `cli:"nlines" dft:"0"`
    Npoints      int               `cli:"n,npoint" dft:"4"`
    Oversamp     float64           `cli:"o,oversamp" dft:"2.0"`
    MaxRad       float64           `cli:"m,maxRadious" dft:"0.0"`
    InterpolMode InterpolationMode `cli:"int,interpol dft:"NearestNeighbour"`
    FlipInput    bool              `cli:"flipIn"`
    FlipOutput   bool              `cli:"flipOut"`
    Order        int               `cli:"r,order" dft:"5"`
}

func (co *CodeOpt) SetCli(c *Cli) {
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



var g2r = Gamma.Must("geocode")

func (l Lookup) geo2radar(in, out IDatFile, opt CodeOpt) (err error) {
    var ferr = merr.Make("Lookup.geo2radar")
    
    lrIn, lrOut := opt.Parse()
    
    if err = opt.RngAzi.Check(); err != nil {
        return ferr.Wrap(err)
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
        return ferr.Wrap(ModeError{name: "interpolation option", got: intm})
    }
    
    dt, dtype := 0, in.Dtype()
    
    switch dtype {
    case Float:
        dt = 0
    case FloatCpx:
        dt = 1
    case Raster:
        dt = 2
    case UChar:
        dt = 3
    case Short:
        dt = 4
    case ShortCpx:
        dt = 5
    case Double:
        dt = 6
    default:
        return ferr.Wrap(WrongTypeError{DType: dtype, kind: "geo2radar"})
    }
    
    
    _, err = g2r(l.Dat, in.Datfile(), in.Rng(),
                 out.Datfile(), out.Rng(),
                 opt.Nlines, interp, dt, lrIn, lrOut, opt.Oversamp,
                 opt.MaxRad, opt.Npoints)
    
    if err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

var r2g = Gamma.Must("geocode_back")

func (l Lookup) radar2geo(in, out IDatFile, opt CodeOpt) (err error) {
    var ferr = merr.Make("Lookup.radar2geo")

    lrIn, lrOut := opt.Parse()
    
    if err = opt.RngAzi.Check(); err != nil {
        return ferr.Wrap(err)
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
            return ferr.Wrap(ModeError{name: "interpolation option", got: intm})
        }
    }
    
    
    dt, dtype := 0, in.Dtype()
    
    switch dtype {
    case Float:
        dt = 0
    case FloatCpx:
        dt = 1
    case Raster:
        dt = 2
    case UChar:
        dt = 3
    case Short:
        dt = 4
    case Double:
        dt = 5
    default:
        err = ferr.Wrap(WrongTypeError{DType: dtype, kind: "radar2geo"})
        return
    }
    
    _, err = r2g(in.Datfile(), in.Rng(), l.Dat,
                 out.Datfile(), out.Rng(),
                 opt.Nlines, interp, dt, lrIn, lrOut, opt.Order)
    
    if err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}
