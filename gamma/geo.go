package gamma

import (
    "log"
    "os"
    fp "path/filepath"
    str "strings"
    conv "strconv"
)

type (
    DEM struct {
        dataFile
        lookup, lookupOld, lsMap, incidence, resolution, offnadir string
    }
    
    Geo2RadarOpt struct {
        width, nlines, dtype int
        interpolMode InterpolationMode
    }
    
    Radar2GeoOpt struct {
        Geo2RadarOpt
        flipInput, flipOutput bool
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
        hgt, simSar, zenith, orient, inc, pix, psi, lsMap, diffPar,
        offs, offsets, ccp, coffs, coffsets, sigma0, gamma0 string
    }
    
    GeoPlotOpt struct {
        rasArgs
        startHgt, startPwr int
        cycle float64
    }
    
    Geocoder struct {
        geocoding
        mli MLI
        outDir string
    }
    
    GeoMeta struct {
        dem, demOrig DEM
        geo Geocode
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

func NewDEM(dat, par, lookup, lookupOld string) (ret DEM, err error) {
    ret.dataFile, err = NewDataFile(par, dat, "par")
    
    if err != nil {
        err = Handle(err, "failed to create DEM struct")
        return
    }
    
    ret.lookup, ret.lookupOld = lookup, lookupOld
    ret.files = []string{dat, par, lookup, lookupOld}

    return ret, nil    
}

func (dem DEM) Rng() (int, error) {
    return dem.Int("width")
}

func (dem DEM) Azi() (int, error) {
    return dem.Int("nlines")
}

var g2r = Gamma.must("geocode")

func (dem *DEM) geo2radar(infile, outfile string, opt Geo2RadarOpt) error {
    if opt.width == 0 {
        return Handle(nil, "datafile width must be specified")
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
    
    rng, err := dem.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive DEM width")
    }
    
    _, err = g2r(dem.lookup, infile, rng, outfile, opt.width, opt.nlines,
                 interp, opt.dtype)
    return err
}

var r2g = Gamma.must("geocode_back")

func (dem *DEM) radar2geo(infile, outfile string, opt Radar2GeoOpt) error {
    lrIn, lrOut := 1, 1
    
    if opt.flipInput {
        lrIn = -1
    }
    
    if opt.flipOutput {
        lrOut = -1
    }
    
    if opt.order == 0 {
        opt.order = 5
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
    
    rng, err := dem.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive DEM width")
    }
    
    _, err = r2g(infile, rng, dem.lookup, outfile, opt.width, opt.nlines,
                 interp, opt.dtype, lrIn, lrOut, opt.order)
    return err
}

var coord2sarpix = Gamma.must("coord_to_sarpix")

func (ll LatLon) ToRadar(mpar, hgt, diffPar string) (ret RngAzi, err error) {
    const par = "corrected SLC/MLI range, azimuth pixel (int)"
    
    out, err := coord2sarpix(mpar, "-", ll.Lat, ll.Lon, hgt, diffPar)
    
    if err != nil {
        err = Handle(err, "failed to retreive radar coordinates")
        return
    }
    
    params := FromString(out, ":")
    
    line, err := params.Par(par)
    
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
        err = Handle(err, "failed to convert string '%s' to int", split[0])
        return
    }
    
    ret.Azi, err = conv.Atoi(split[1])

    if err != nil {
        err = Handle(err, "failed to convert string '%s' to int", split[0])
        return
    }

    return ret, nil
}

var rasmph = Gamma.must("rasmph")

func (d *DEM) Raster(mode DEMPlot, opt rasArgs) error {
    err := opt.Parse(d)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    switch mode {
    case Lookup:
        opt.disArgs.Datfile = d.lookup
        opt.ImgFmt = "FCOMPLEX"
        opt.Cmd = "mph"
    case Dem:
        opt.disArgs.Datfile = d.dat
        opt.Cmd = "hgt"
    default:
        return Handle(nil, "unrecognized plot mode")
    }
    
    return Raster(dem, opt, "")
}


func NewGeocode(dat, par string) (ret Geocode, err error) {
    ret.MLI, err = NewMLI(dat, par)
    return
} 

func (opt *GeoPlotOpt) Parse(d DataFile) error {
    err := opt.rasArgs.Parse(d)
    
    if err != nil {
        return Handle(err, "failed to parse plotting options")
    }
    
    if opt.cycle == 0 {
        opt.cycle = 160.0
    }
    
    return nil
}

var rashgt = Gamma.must("rashgt")

func (geo *Geocode) Raster(opt GeoPlotOpt) error {
    err := opt.Parse(geo)
    
    if err != nil {
        return Handle(err, "failed to parse plot arguments")
    }
    
    _, err = rashgt(geo.hgt, geo.MLI.dat, opt.Rng, opt.startHgt, opt.startPwr,
                    opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.cycle, opt.Scale,
                    opt.Exp, opt.LR, opt.raster)
    return err
}


var (
    createDiffPar = Gamma.must("create_diff_par")
    vrt2dem = Gamma.must("vrt2dem")
    gcMap = Gamma.must("gc_map2")
    pixelArea = Gamma.must("pixel_area")
    offsetPwrm = Gamma.must("offset_pwrm")
    offsetFitm = Gamma.must("offset_fitm")
    gcMapFine = Gamma.must("gc_map_fine")
)

func (g* Geocoder) Run() (ret GeoMeta, err error) {
    out := g.outDir
    
    demdir := fp.Join(out, "dem")
    geodir := fp.Join(out, "geo")
    
    demOrig, err := NewDEM(fp.Join(demdir, "srtm.dem"), "", "", "")
    
    if err != nil {
        err = Handle(err, "failed to create DEM struct")
        return
    }
    
    vrtPath := g.DEMPath
    
    if len(vrtPath) == 0 {
        err = Handle(nil, "path to vrt files not specified")
        return
    }
    
    oversamp := g.DEMOverSampling
    
    if oversamp.Lat == 0.0 {
        oversamp.Lat = 2.0
    }
    
    if oversamp.Lon == 0.0 {
        oversamp.Lon = 2.0
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
    
    
    ex, err := Exist(demOrig.dat)
    
    if err != nil {
        err = Handle(err, "failed to check whether original DEM exists")
        return
    }
    
    mli := g.mli
    
    if !ex {
        log.Printf("Creating DEM from %s\n", vrtPath)
        
        _, err = vrt2dem(vrtPath, mli.par, demOrig.dat, demOrig.par, 2, nil)
        
        if err != nil {
            err = Handle(err, "failed to create DEM from vrt file")
            return
        }
    } else {
        log.Println("DEM already imported.")
    }
            
    mra := RngAzi{}
    
    mra.Rng, err = mli.Rng()
    
    if err != nil {
        err = Handle(err, "failed to retreive reference mli range samples")
        return
    }
    
    mra.Azi, err = mli.Azi()
    
    if err != nil {
        err = Handle(err, "failed to retreive reference mli azimuth lines")
        return
    }
    
    offsetWin := g.OffsetWindows
    
    Patch := RngAzi{
        Rng: int(float64(mra.Rng) / float64(offsetWin.Rng) + float64(overlap.Rng) / 2),
        Azi: int(float64(mra.Azi) / float64(offsetWin.Azi) + float64(overlap.Azi) / 2),
    }
    
    // make sure the number of patches are even
    
    if Patch.Rng % 2 == 1 {
        Patch.Rng += 1
    }
    
    if Patch.Azi % 2 == 1 {
        Patch.Azi += 1
    }
    
    dem, err := NewDEM(fp.Join(demdir, "dem_seg.dem"), "",
        fp.Join(geodir, "lookup"), fp.Join(geodir, "lookup_old"))
    
    if err != nil {
        err = Handle(err, "failed to create DEM struct")
        return
    }
    
    
    geo := Geocode{
        hgt     : fp.Join(geodir, "hgt"),
        lsMap   : fp.Join(geodir, "lsMap"),
        simSar  : fp.Join(geodir, "sim_sar"),
        zenith  : fp.Join(geodir, "zenith"),
        orient  : fp.Join(geodir, "orient"),
        inc     : fp.Join(geodir, "inc"),
        pix     : fp.Join(geodir, "pix"),
        psi     : fp.Join(geodir, "psi"),
        diffPar : fp.Join(geodir, "diff_par"),
        offs    : fp.Join(geodir, "offs"),
        offsets : fp.Join(geodir, "offsets"),
        ccp     : fp.Join(geodir, "ccp"),
        coffs   : fp.Join(geodir, "coffs"),
        coffsets: fp.Join(geodir, "coffsets"),
    }
    
    geo.MLI = mli
    
    ex1, err := Exist(dem.lookup)
    
    if err != nil {
        err = Handle(err, "failed to check whether lookup table exists")
        return
    }
    
    ex2, err := Exist(dem.par)
    
    if err != nil {
        err = Handle(err, "failed to check whether DEM parameter exists")
        return
    }
    
    if !ex1 && !ex2 {
        log.Println("Calculating initial lookup table.")
        
        _, err = gcMap(mli.par, demOrig.par, demOrig.dat, dem.par, dem.dat,
                       dem.lookup, oversamp.Lat, oversamp.Lon, dem.lsMap,
                       geo.lsMap, dem.incidence, dem.resolution,
                       dem.offnadir, g.RngOversamp, Standard, NoMask,
                       g.nPixel, geo.diffPar, Actual)
        
        /* old
        _, err = gcMap(mli.par, nil, demOrig.par, demOrig.dat, dem.par, dem.dat,
                       dem.lookup, oversamp.Lat, oversamp.Lon, geo.simSar,
                       geo.zenith, geo.orient, geo.inc, geo.psi, geo.pix,
                       geo.lsMap, 8, 2)
        */
        
        if err != nil {
            err = Handle(err, "gc_map2 failed")
            return
        }      
    } else {
        log.Println("Initial lookup table already created.")
    }
    
    dra := RngAzi{}
    
    dra.Rng, err = dem.Rng()
    
    if err != nil {
        err = Handle(err, "failed to retreive DEM range samples")
        return
    }
    
    dra.Azi, err = dem.Azi()
    
    if err != nil {
        err = Handle(err, "failed to retreive DEM azimuth lines")
        return
    }
    
    _, err = pixelArea(mli.par, dem.par, dem.dat, dem.lookup, dem.lsMap,
                       dem.incidence, geo.sigma0, geo.gamma0, g.AreaFactor)
    
    if err != nil {
        err = Handle(err, "pixel area failed")
        return
    }
    
    
    _, err = createDiffPar(mli.par, nil, geo.diffPar, SLC_MLI, NonInter)
    
    if err != nil {
        err = Handle(err, "create_diff_par failed")
        return
    }
    
    log.Println("Refining lookup table.")
    
    if itr >= 1 {
        log.Println("ITERATING OFFSET REFINEMENT.")
        
        for ii := 0; ii < itr; ii++ {
            log.Printf("ITERATION %d / %d\n", ii + 1, itr)
            
            err = os.Remove(geo.diffPar)
            
            if err != nil {
                err = Handle(err, "failed to remove file '%s'", geo.diffPar)
                return
            }

            // copy previous lookup table
            err = os.Rename(dem.lookup, dem.lookupOld)
            
            if err != nil {
                err = Handle(err, "failed to move lookup file '%s'",
                    dem.lookup)
                return
            }
            
            _, err = createDiffPar(mli.par, nil, geo.diffPar,
                                   SLC_MLI, NonInter)
            
            if err != nil {
                err = Handle(err, "create_diff_par failed")
                return
            }
            
            _, err = offsetPwrm(geo.sigma0, mli.dat, geo.diffPar, geo.offs,
                                geo.ccp, Patch.Rng, Patch.Azi, geo.offsets,
                                g.MLIOversamp, offsetWin.Rng, offsetWin.Azi,
                                g.CCThresh, g.LanczosOrder, g.BandwithFrac)
            
            if err != nil {
                err = Handle(err, "offset_pwrm failed")
                return
            }
            
            _, err = offsetFitm(geo.offs, geo.ccp, geo.diffPar, geo.coffs,
                                geo.coffsets, g.CCThresh, npoly, NonInter)
            
            
            if err != nil {
                err = Handle(err, "offset_fitm failed")
                return
            }

            // update previous lookup table
            // TODO: magic number 1
            _, err = gcMapFine(dem.lookupOld, dra.Rng, geo.diffPar,
                               dem.lookup, 1)
            
            if err != nil {
                err = Handle(err, "gc_map_fine failed")
                return
            }

            // create new simulated ampliutides with the new lookup table
            // TODO: magic number: 20
            _, err = pixelArea(mli.par, dem.par, dem.dat, dem.lookup, geo.lsMap,
                               geo.inc, geo.sigma0, geo.gamma0, 20)
            
            if err != nil {
                err = Handle(err, "pixel_area failed")
                return
            }

        }
        log.Println("ITERATION DONE.")
    }
    
    
    ret = GeoMeta{
        geo: geo,
        demOrig: demOrig,
        dem: dem,
    }
    
    return ret, nil
}


