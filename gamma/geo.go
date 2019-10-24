package gamma

import (
    "log"
    "fmt"
    "os"
    fp "path/filepath"
    str "strings"
    conv "strconv"
)

type (
    DEM struct {
        dataFile
        Lookup, LookupOld string
    }
    
    CodeOpt struct {
        inWidth int
        outWidth int
        nlines int
        npoints int
        dtype string
        oversamp float64
        maxRad float64
        interpolMode InterpolationMode
        flipInput bool
        flipOutput bool
        order int
    }
    
    InterpolationMode int
    DEMPlot int
    LayoverScaling int
    MaskingMode int
    RefenceMode int
    ParFileType int
    InteractMode int
    
    
    Geocode struct {
        MLI
        Hgt, SimSar, Zenith, Orient, Inc, Pix, Psi, LsMap, DiffPar,
        Offs, Offsets, Ccp, Coffs, Coffsets, Sigma0, Gamma0 string
    }
    
    GeoPlotOpt struct {
        RasArgs
    }
    
    GeoMeta struct {
        Dem, DemOrig DEM
        Geo Geocode
    }
)   

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

const (
    Dem DEMPlot = iota
    Lookup
)

const (
    Standard LayoverScaling = iota
    VisOpt
)

const (
    NoMask MaskingMode = iota
    MaskOutside
    MaskFull
)

const(
    Actual RefenceMode = iota
    Simulated
)

const (
    Offset ParFileType = iota
    SLC_MLI
    Elevation
)

const (
    NonInter InteractMode = iota
    Inter
)

// TODO: check datatype
func NewDEM(dat, par, lookup, lookupOld string) (ret DEM, err error) {
    if ret.dataFile, err = Newdatafile(dat, par); err != nil {
        err = DataCreateErr.Wrap(err, "DEM")
        //err = Handle(err, "failed to create DEM struct")
        return
    }
    
    if ret.Rng, err = ret.rng(); err != nil {
        err = RngError.Wrap(err, par)
        //err = Handle(err, "failed to retreive range samples from '%s'", par)
        return
    }
    
    if ret.Azi, err = ret.azi(); err != nil {
        err = AziError.Wrap(err, par)
        //err = Handle(err, "failed to retreive azimuth lines from '%s'", par)
        return
    }
    
    ret.Lookup, ret.LookupOld = lookup, lookupOld

    return ret, nil    
}

func (dem DEM) rng() (int, error) {
    return dem.Int("width", 0)
}

func (dem DEM) azi() (int, error) {
    return dem.Int("nlines", 0)
}

func (opt *CodeOpt) Parse() (lrIn int, lrOut int, err error) {
    lrIn, lrOut = 1, 1
    
    if opt.flipInput {
        lrIn = -1
    }
    
    if opt.flipOutput {
        lrOut = -1
    }
    
    if opt.order == 0 {
        opt.order = 5
    }
    
    if opt.inWidth == 0 {
        err = Handle(nil, "infile width must be specified")
        return
    }
    
    if opt.outWidth == 0 {
        err = Handle(nil, "outfile width must be specified")
        return
    }
    
    if opt.oversamp == 0.0 {
        opt.oversamp = 2.0
    }
    
    if opt.maxRad == 0.0 {
        opt.maxRad = 4 * opt.oversamp
    }
    
    if opt.npoints == 0 {
        opt.npoints = 4
    }
    
    return lrIn, lrOut, nil
}

var g2r = Gamma.Must("geocode")

func (dem *DEM) geo2radar(infile, outfile string, opt CodeOpt) error {
    lrIn, lrOut, err := opt.Parse()
    
    if err != nil {
        return Handle(err, "failed to parse geo2radar arguments")
    }
    
    interp := 0

    switch opt.interpolMode {
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
        return Handle(nil, "unrecognized interpolation option")
    }
    
    dtype := 0
    
    switch opt.dtype {
        case "FLOAT":
            dtype = 0
        case "FCOMPLEX":
            dtype = 1
        case "SUN", "raster", "BMP", "TIFF":
            dtype = 2
        case "UNSIGNED CHAR":
            dtype = 3
        case "SHORT":
            dtype = 4
        case "SCOMPLEX":
            dtype = 5
        case "DOUBLE":
            dtype = 6
        default:
            return Handle(nil, "unrecognized data format: %s", opt.dtype)
    }
    
    // rng, err := dem.Rng()
    
    // if err != nil {
    //     return Handle(err, "failed to retreive DEM width")
    // }
    
    _, err = g2r(dem.Lookup, infile, opt.inWidth, outfile, opt.outWidth,
                 opt.nlines, interp, dtype, lrIn, lrOut, opt.oversamp,
                 opt.maxRad, opt.npoints)
    
    return err
}

var r2g = Gamma.Must("geocode_back")

func (dem DEM) radar2geo(infile, outfile string, opt CodeOpt) error {
    lrIn, lrOut, err := opt.Parse()
    
    if err != nil {
        return Handle(err, "failed to parse radar2geo arguments")
    }
    
    var interp int
    
    // default interpolation mode
    if opt.interpolMode == NearestNeighbour {
        interp = 1
    } else {
        switch opt.interpolMode {
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
            return Handle(nil, "unrecognized interpolation option")
        }
    }
    
    dtype := 0
    
    switch opt.dtype {
        case "FLOAT":
            dtype = 0
        case "FCOMPLEX":
            dtype = 1
        case "SUN", "raster", "BMP", "TIFF":
            dtype = 2
        case "UNSIGNED CHAR":
            dtype = 3
        case "SHORT":
            dtype = 4
        case "DOUBLE":
            dtype = 5
        default:
            return Handle(nil, "unrecognized data format: %s", opt.dtype)
    }
    
    // TODO: use this if opt.inWidth == 0?
    // rng, err := dem.Rng()
    
    // if err != nil {
    //     return Handle(err, "failed to retreive DEM width")
    // }
    
    if opt.order == 0 {
        opt.order = 5
    }
    
    _, err = r2g(infile, opt.inWidth, dem.Lookup, outfile, opt.outWidth,
                 opt.nlines, interp, dtype, lrIn, lrOut, opt.order)
    
    return err
}

var coord2sarpix = Gamma.Must("coord_to_sarpix")

func (ll LatLon) ToRadar(mpar, hgt, diffPar string) (ret RngAzi, err error) {
    const par = "corrected SLC/MLI range, azimuth pixel (int)"
    
    out, err := coord2sarpix(mpar, "-", ll.Lat, ll.Lon, hgt, diffPar)
    
    if err != nil {
        err = Handle(err, "failed to retreive radar coordinates")
        return
    }
    
    params := FromString(out, ":")
    
    line, err := params.Param(par)
    
    if err != nil {
        err = Handle(err, "failed to retreive range, azimuth")
        return
    }
    
    split := str.Split(line, " ")
    
    if len(split) < 2 {
        err = Handle(nil, "split to retreive range, azimuth failed")
        return
    }
    
    ret.Rng, err = conv.Atoi(split[0])

    if err != nil {
        err = ParseIntErr.Wrap(err, split[0])
        //err = Handle(err, "failed to convert string '%s' to int", split[0])
        return
    }
    
    ret.Azi, err = conv.Atoi(split[1])

    if err != nil {
        err = ParseIntErr.Wrap(err, split[1])
        //err = Handle(err, "failed to convert string '%s' to int", split[0])
        return
    }

    return ret, nil
}

func (d DEM) Raster(opt RasArgs) error {
    opt.DisArgs.Datfile = d.Dat
    opt.ImgFmt = "FLOAT"
    
    err := opt.Parse(d)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    /*
    switch mode {
    case Lookup:
        opt.DisArgs.Datfile = d.Lookup
        opt.ImgFmt = "FCOMPLEX"
        
        err := opt.Parse(d)
        
        if err != nil {
            return Handle(err, "failed to parse raster options")
        }
        
        return rasmph(opt)
    case Dem:
    default:
        return Handle(nil, "unrecognized plot mode")
    }
    */
    
    return nil
}


func NewGeocode(dat, par string) (ret Geocode, err error) {
    ret.MLI, err = NewMLI(dat, par)
    return
} 

func (opt *GeoPlotOpt) Parse(d DataFile) error {
    err := opt.RasArgs.Parse(d)
    
    if err != nil {
        return Handle(err, "failed to parse plotting options")
    }
    
    if opt.StartPwr == 0 {
        opt.StartPwr = 1
    }
    
    if opt.StartHgt == 0 {
        opt.StartHgt = 1
    }
    
    
    return nil
}

var rashgt = Gamma.Must("rashgt")

func (geo Geocode) Raster(opt RasArgs) error {
    opt.Raster = fmt.Sprintf("%s.%s", geo.Hgt, Settings.RasExt)
    
    err := opt.Parse(geo)
    
    if err != nil {
        return Handle(err, "failed to parse plot arguments")
    }
    
    _, err = rashgt(geo.Hgt, geo.MLI.Dat, opt.Rng, opt.StartHgt, opt.StartPwr,
                    opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.Cycle, opt.Scale,
                    opt.Exp, opt.LR, opt.Raster)
    return err
}


var (
    createDiffPar = Gamma.Must("create_diff_par")
    vrt2dem = Gamma.Must("vrt2dem")
    //gcMap = Gamma.Must("gc_map2")
    gcMap = Gamma.Must("gc_map")
    pixelArea = Gamma.Must("pixel_area")
    offsetPwrm = Gamma.Must("offset_pwrm")
    offsetFitm = Gamma.Must("offset_fitm")
    gcMapFine = Gamma.Must("gc_map_fine")
)

func (g* GeocodeOpt) Run(outDir string) (ret GeoMeta, err error) {
    geodir := fp.Join(outDir, "geo")
    
    err = os.MkdirAll(geodir, os.ModePerm)
    
    if err != nil {
        err = DirCreateErr.Wrap(err, geodir)
        //err = Handle(err, "failed to create directory '%s'!", geodir)
        return
    }
    
    demOrig, err := NewDEM(fp.Join(geodir, "srtm.dem"), "", "", "")
    
    if err != nil {
        err = DataCreateErr.Wrap(err, "DEM")
        //err = Handle(err, "failed to create DEM struct")
        return
    }
    
    vrtPath := g.DEMPath
    
    if len(vrtPath) == 0 {
        err = Handle(nil, "path to vrt files not specified")
        return
    }
    
    overlap := g.DEMOverlap
    
    if overlap.Rng == 0 {
        overlap.Rng = 100
    }
    
    if overlap.Azi == 0 {
        overlap.Azi = 100
    }

    npoly, itr := g.NPoly, g.Iter
    
    if npoly == 0 {
        npoly = 4
    }
    
    
    ex, err := Exist(demOrig.Dat)
    
    if err != nil {
        err = Handle(err, "failed to check whether original DEM exists")
        return
    }
    
    mli, err := NewMLI(g.Master.Dat, g.Master.Par)
    
    if err != nil {
        err = Handle(err, "failed to parse master MLI file")
        return
    }
    
    if !ex {
        log.Printf("Creating DEM from %s\n", vrtPath)
        
        // magic number 2 = add interpolated geoid offset
        _, err = vrt2dem(vrtPath, mli.Par, demOrig.Dat, demOrig.Par, 2, "-")
        
        if err != nil {
            err = Handle(err, "failed to create DEM from vrt file")
            return
        }
    } else {
        log.Println("DEM already imported.")
    }
    
    mra := mli.RngAzi
    offsetWin := g.OffsetWindows
    
    Patch := RngAzi{
        Rng: int(float64(mra.Rng) / float64(offsetWin.Rng) +
             float64(overlap.Rng) / 2),
        
        Azi: int(float64(mra.Azi) / float64(offsetWin.Azi) +
             float64(overlap.Azi) / 2),
    }
    
    // make sure the number of patches are even
    
    if Patch.Rng % 2 == 1 {
        Patch.Rng += 1
    }
    
    if Patch.Azi % 2 == 1 {
        Patch.Azi += 1
    }
    
    dem, err := NewDEM(fp.Join(geodir, "dem_seg.dem"), "",
        fp.Join(geodir, "lookup"), fp.Join(geodir, "lookup_old"))
    
    if err != nil {
        err = DataCreateErr.Wrap(err, "DEM")
        //err = Handle(err, "failed to create DEM struct")
        return
    }
    
    
    geo := Geocode{
        Hgt     : fp.Join(geodir, "hgt"),
        Sigma0     : fp.Join(geodir, "sigma0"),
        Gamma0     : fp.Join(geodir, "gamma0"),
        LsMap   : fp.Join(geodir, "lsMap"),
        SimSar  : fp.Join(geodir, "sim_sar"),
        Zenith  : fp.Join(geodir, "zenith"),
        Orient  : fp.Join(geodir, "orient"),
        Inc     : fp.Join(geodir, "inc"),
        Pix     : fp.Join(geodir, "pix"),
        Psi     : fp.Join(geodir, "psi"),
        DiffPar : fp.Join(geodir, "diff_par"),
        Offs    : fp.Join(geodir, "offs"),
        Offsets : fp.Join(geodir, "offsets"),
        Ccp     : fp.Join(geodir, "ccp"),
        Coffs   : fp.Join(geodir, "coffs"),
        Coffsets: fp.Join(geodir, "coffsets"),
    }
    
    geo.MLI = mli
    
    ex1, err := Exist(dem.Lookup)
    
    if err != nil {
        err = Handle(err, "failed to check whether lookup table exists")
        return
    }
    
    ex2, err := Exist(dem.Par)
    
    if err != nil {
        err = Handle(err, "failed to check whether DEM parameter exists")
        return
    }
    
    if !ex1 && !ex2 {
        log.Println("Calculating initial lookup table.")
        
        oversamp := g.DEMOverSampling
        
        if oversamp.Lat < 1.0 {
            oversamp.Lat = 2.0
        }
        
        if oversamp.Lon < 1.0 {
            oversamp.Lon = 2.0
        }
        
        if g.RngOversamp < 1.0 {
            g.RngOversamp = 2.0
        }
        
        /*
        _, err = gcMap(mli.par, demOrig.par, demOrig.dat, dem.par, dem.dat,
                       dem.lookup, oversamp.Lat, oversamp.Lon, demOrig.lsMap,
                       geo.lsMap, demOrig.incidence, demOrig.resolution,
                       demOrig.offnadir, g.RngOversamp, Standard, NoMask,
                       g.nPixel, "-", Actual)
        */
        
        _, err = gcMap(mli.Par, nil, demOrig.Par, demOrig.Dat, dem.Par, dem.Dat,
                       dem.Lookup, oversamp.Lat, oversamp.Lon, geo.SimSar,
                       geo.Zenith, geo.Orient, geo.Inc, geo.Psi, geo.Pix,
                       geo.LsMap, g.nPixel, 2, g.RngOversamp)
        
        if err != nil {
            err = Handle(err, "gc_map failed")
            return
        }      
    } else {
        log.Println("Initial lookup table already created.")
    }
    
    dra := dem.RngAzi
    
    _, err = pixelArea(mli.Par, dem.Par, dem.Dat, dem.Lookup, geo.LsMap,
                       geo.Inc, geo.Sigma0, geo.Gamma0, g.AreaFactor)
    
    if err != nil {
        err = Handle(err, "pixel area failed")
        return
    }
    
    _, err = createDiffPar(mli.Par, nil, geo.DiffPar, SLC_MLI, NonInter)
    
    if err != nil {
        err = Handle(err, "create_diff_par failed")
        return
    }
    
    log.Println("Refining lookup table.")
    
    if itr >= 1 {
        log.Println("ITERATING OFFSET REFINEMENT.")
        
        for ii := 0; ii < itr; ii++ {
            log.Printf("ITERATION %d / %d\n", ii + 1, itr)
            
            err = os.Remove(geo.DiffPar)
            
            if err != nil {
                err = Handle(err, "failed to remove file '%s'", geo.DiffPar)
                return
            }

            // copy previous lookup table
            err = os.Rename(dem.Lookup, dem.LookupOld)
            
            if err != nil {
                err = Handle(err, "failed to move lookup file '%s'",
                    dem.Lookup)
                return
            }
            
            _, err = createDiffPar(mli.Par, nil, geo.DiffPar,
                                   SLC_MLI, NonInter)
            
            if err != nil {
                err = Handle(err, "create_diff_par failed")
                return
            }
            
            _, err = offsetPwrm(geo.Sigma0, mli.Dat, geo.DiffPar, geo.Offs,
                                geo.Ccp, Patch.Rng, Patch.Azi, geo.Offsets,
                                g.MLIOversamp, offsetWin.Rng, offsetWin.Azi,
                                g.CCThresh, g.LanczosOrder, g.BandwithFrac)
            
            if err != nil {
                err = Handle(err, "offset_pwrm failed")
                return
            }
            
            _, err = offsetFitm(geo.Offs, geo.Ccp, geo.DiffPar, geo.Coffs,
                                geo.Coffsets, g.CCThresh, npoly, NonInter)
            
            
            if err != nil {
                err = Handle(err, "offset_fitm failed")
                return
            }

            // update previous lookup table
            // TODO: magic number 1
            _, err = gcMapFine(dem.LookupOld, dra.Rng, geo.DiffPar,
                               dem.Lookup, 1)
            
            if err != nil {
                err = Handle(err, "gc_map_fine failed")
                return
            }

            // create new simulated ampliutides with the new lookup table
            _, err = pixelArea(mli.Par, dem.Par, dem.Dat, dem.Lookup, geo.LsMap,
                               geo.Inc, geo.Sigma0, geo.Gamma0, g.AreaFactor)
            
            if err != nil {
                err = Handle(err, "pixel_area failed")
                return
            }

        }
        log.Println("ITERATION DONE.")
    }
    
    
    ret = GeoMeta{
        Geo: geo,
        DemOrig: demOrig,
        Dem: dem,
    }
    
    return ret, nil
}


