package gamma

import (
    "log"
    "fmt"
    "os"
    "path/filepath"
    
    "./data"
)


type DEM struct {
    datafile.DatFile
}

func (d *DEM) Decode(s string) (err error) {
    return LoadJson(s, d)
}


type Lookup struct {
    datafile.DatFile
}

func (l *Lookup) Decode(s string) (err error) {
    return LoadJson(s, l)
}

func NewDEM(dat, par string) (d DEM, err error) {
    var ferr = merr.Make("NewDEM")
    
    if d.DatParFile, err = NewDatParFile(dat, par, "par", Float);
       err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (d DEM) NewLookup(path string) (l Lookup) {
    l.Dat = path
    l.Ra = d.Ra
    l.DType = FloatCpx
    return
}

func (dem DEM) ParseRng() (i int, err error) {
    var ferr = merr.Make("DEM.ParseRng")
    
    if i, err = dem.Int("width", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (dem DEM) ParseAzi() (i int, err error) {
    var ferr = merr.Make("DEM.ParseAzi")
    
    if i, err = dem.Int("nlines", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (d DEM) Raster(opt RasArgs) (err error) {
    var ferr = merr.Make("DEM.Raster")
    opt.Mode = Power
    opt.Parse(d)
    
    if err = d.DatFile.Raster(opt); err != nil {
        err = ferr.Wrap(err)
    }
    
    return nil
}

func (l Lookup) Raster(opt RasArgs) (err error) {
    var ferr = merr.Make("Lookup.Raster")
    
    opt.Mode = MagPhase
    opt.Parse(l)

    if err = l.DatFile.Raster(opt); err != nil {
        err = ferr.Wrap(err)
    }
    
    return nil
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

var coord2sarpix = Gamma.Must("coord_to_sarpix")

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

type Hgt struct {
    DatFile
}

func (h *Hgt) Decode(s string) (err error) {
    return LoadJson(s, h)
}

func (h Hgt) Raster(opt RasArgs) (err error) {
    var ferr = merr.Make("Hgt.Raster")
    
    opt.Mode = Height
    opt.Parse(h)
    
    if err = h.DatFile.Raster(opt); err != nil {
        err = ferr.Wrap(err)
    }
    
    return nil
}

type Geocode struct {
    MLI
    DiffPar, Offs, Offsets, Ccp, Coffs, Coffsets, Sigma0, Gamma0 string
}

type (
    LayoverScaling int
    MaskingMode int
    RefenceMode int
    ParFileType int
    InteractMode int
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

func (g* GeocodeOpt) Run(outDir string) (err error) {
    var ferr = merr.Make("GeocodeOpt")
    
    geodir := filepath.Join(outDir, "geo")
    
    if err = Mkdir(geodir); err != nil {
        return ferr.Wrap(err)
    }
    
    var demOrig DEM
    if demOrig, err = NewDEM(filepath.Join(geodir, "srtm.dem"), "");
       err != nil {
        return ferr.Wrap(err)
    }
    
    vrtPath := g.DEMPath
    
    if len(vrtPath) == 0 {
        return ferr.Wrap(fmt.Errorf("path to vrt files not specified"))
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
    
    
    
    var ex bool
    if ex, err = Exist(demOrig.Dat); err != nil {
        return ferr.WrapFmt(err,
            "failed to check whether original DEM exists")
    }
    
    var mli MLI
    if mli, err = NewMLI(g.Master.Dat, g.Master.Par); err != nil {
        return ferr.WrapFmt(err, "failed to parse master MLI file")
    }
    
    if !ex {
        log.Printf("Creating DEM from %s\n", vrtPath)
        
        // magic number 2 = add interpolated geoid offset
        _, err = vrt2dem(vrtPath, mli.Par, demOrig.Dat, demOrig.Par, 2, "-")
        
        if err != nil {
            return ferr.WrapFmt(err, "failed to create DEM from vrt file")
        }
    } else {
        log.Println("DEM already imported.")
    }
    
    mra := mli.Ra
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
    
    var dem DEM
    if dem, err = NewDEM(filepath.Join(geodir, "dem_seg.dem"), "");
       err != nil {
        return ferr.Wrap(err)
    }
    
    geo := Geocode{
        Offs    : filepath.Join(geodir, "offs"),
        Offsets : filepath.Join(geodir, "offsets"),
        Ccp     : filepath.Join(geodir, "ccp"),
        Coffs   : filepath.Join(geodir, "coffs"),
        Coffsets: filepath.Join(geodir, "coffsets"),
        DiffPar : filepath.Join(geodir, "diff_par"),
        MLI     : mli,
    }
    
    var sigma0, gamma0, lsMap, simSar, zenith, orient, inc, pix, proj DatFile
    
    if sigma0, err =  mli.Like(filepath.Join(geodir, "sigma0"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if gamma0, err =  mli.Like(filepath.Join(geodir, "gamma0"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    // datatype of lsmap?
    if lsMap, err =  mli.Like(filepath.Join(geodir, "lsmap"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if simSar, err =  mli.Like(filepath.Join(geodir, "sim_sar"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if zenith, err =  mli.Like(filepath.Join(geodir, "zenith"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if orient, err =  mli.Like(filepath.Join(geodir, "orient"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if inc, err =  mli.Like(filepath.Join(geodir, "inclination"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if proj, err =  mli.Like(filepath.Join(geodir, "projection"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    if pix, err =  mli.Like(filepath.Join(geodir, "pixel_area"), Float);
       err != nil {
        return ferr.Wrap(err)
    }
    
    var lookup Lookup
    if lookup.DatFile, err = dem.Like(filepath.Join(geodir, "lookup"), FloatCpx);
       err != nil {
        return ferr.Wrap(err)
    }
    
    var lookupOld string
    if lookupOld, err = TmpFile(""); err != nil {
        return ferr.Wrap(err)
    }
    
    var ex1 bool
    if ex1, err = Exist(lookup.Dat); err != nil {
        return ferr.WrapFmt(err,
            "failed to check whether lookup table exists")
        return
    }
    
    
    var ex2 bool
    if ex2, err = Exist(dem.Par); err != nil {
        return ferr.WrapFmt(err,
            "failed to check whether DEM parameter exists")
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
                       lookup.Dat, oversamp.Lat, oversamp.Lon, simSar.Dat,
                       zenith.Dat, orient.Dat, inc.Dat, proj.Dat, pix.Dat,
                       lsMap.Dat, g.nPixel, 2, g.RngOversamp)
        
        if err != nil {
            return ferr.Wrap(err)
        }      
    } else {
        log.Println("Initial lookup table already created.")
    }
    
    dra := dem.Ra
    
    _, err = pixelArea(mli.Par, dem.Par, dem.Dat, lookup.Dat, lsMap.Dat,
                       inc.Dat, sigma0.Dat, gamma0.Dat, g.AreaFactor)
    
    if err != nil {
        return ferr.Wrap(err)
    }
    
    _, err = createDiffPar(mli.Par, nil, geo.DiffPar, SLC_MLI, NonInter)
    
    if err != nil {
        return ferr.Wrap(err)
    }
    
    log.Println("Refining lookup table.")
    
    if itr >= 1 {
        log.Println("ITERATING OFFSET REFINEMENT.")
        
        for ii := 0; ii < itr; ii++ {
            log.Printf("ITERATION %d / %d\n", ii + 1, itr)
            
            if err = os.Remove(geo.DiffPar); err != nil {
                return ferr.WrapFmt(err, "failed to remove file '%s'",
                    geo.DiffPar)
            }

            // copy previous lookup table
            if err = os.Rename(lookup.Dat, lookupOld); err != nil {
                return ferr.WrapFmt(err,
                    "failed to move lookup file '%s'", lookup.Dat)
            }
            
            _, err = createDiffPar(mli.Par, nil, geo.DiffPar,
                                   SLC_MLI, NonInter)
            
            if err != nil {
                return ferr.Wrap(err)
            }
            
            _, err = offsetPwrm(geo.Sigma0, mli.Dat, geo.DiffPar, geo.Offs,
                                geo.Ccp, Patch.Rng, Patch.Azi, geo.Offsets,
                                g.MLIOversamp, offsetWin.Rng, offsetWin.Azi,
                                g.CCThresh, g.LanczosOrder, g.BandwithFrac)
            
            if err != nil {
                return ferr.Wrap(err)
            }
            
            _, err = offsetFitm(geo.Offs, geo.Ccp, geo.DiffPar, geo.Coffs,
                                geo.Coffsets, g.CCThresh, npoly, NonInter)
            
            
            if err != nil {
                return ferr.Wrap(err)
            }

            // update previous lookup table
            // TODO: magic number 1
            _, err = gcMapFine(lookupOld, dra.Rng, geo.DiffPar,
                               lookup.Dat, 1)
            
            if err != nil {
                return ferr.Wrap(err)
            }

            // create new simulated ampliutides with the new lookup table
            _, err = pixelArea(mli.Par, dem.Par, dem.Dat, lookup.Dat, lsMap.Dat,
                               inc.Dat, sigma0.Dat, gamma0.Dat, g.AreaFactor)
            
            if err != nil {
                return ferr.Wrap(err)
            }

        }
        log.Println("ITERATION DONE.")
    }
    
    
    toSave := []IDatFile{
        &dem, &demOrig, &lookup, &sigma0, &gamma0, &lsMap, &simSar, &zenith,
        &orient, &inc, &pix, &proj,
    }
    
    for _, s := range toSave {
        if err = SaveJson(s.Datfile() + ".json", s); err != nil {
            return ferr.Wrap(err)
        }
    }
    
    return nil
}

