package gamma

type (
    DEM struct {
        dataFile
        lookup string
    }
    
    Geo2RadarOpt struct {
        width, nlines, dtype int
        interp geo2radarMode
    }
    
    geo2radarMode int
    radar2geoMode int
    plotMode int
    
    Geocode struct {
        MLI
        hgt, simSar, zenith, orient, inc, pix, psi, lsMap, diffPar, offs,
        ccp, coffs, coffsets string
    }
    
    GeoPlotOpt struct {
        
    }
)   

const (
    Distance geo2radarMode = iota
    NearestNeighbour
    SqaredDistance
    Constant
    Gauss
)

const (
    NearestNeighbour radar2geoMode = iota
)

const (
    Dem plotMode plotMode = iota
    Lookup
)
    

func NewDEM(dat, par, lookup string) (ret DEM, err error) {
    ret.dataFile, err = NewDataFile(par, dat, "par")
    
    if err != nil {
        err = Handle(err, "failed to create DEM struct")
        return
    }
    
    ret.files = []string{dat, par, lookup}

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
    
    rng, err := dem.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive DEM width")
    }
    
    return g2r(dem.lookup, infile, rng, outfile, opt.width, opt.nlines,
               opt.interp, opt.dtype)
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
    
    rng, err := dem.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive DEM width")
    }
    
    return r2g(infile, rng, dem.lookup, outfile, opt.width, opt.nlines,
               opt.interp, opt.dtype, lrIn, lrOut, opt.order)
}

var coord2sarpix = Gamma.must("coord_to_sarpix")

func (ll LatLon) ToRadar(mpar, hgt, diffPar string) (ret RngAzi, err error) {
    const par = "corrected SLC/MLI range, azimuth pixel (int)"
    
    out, err := coord2sarpix(mpar, "-", ll.Lat, ll.Lon, hgt, diffPar)
    
    if err != nil {
        err = Handle(err, "failed to retreive radar coordinates")
        return
    }
    
    par := FromString(out, ":")
    
    line, err := par.Par(par)
    
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

func (dem *DEM) Raster(mode plotMode, opt rasArgs) error {
    switch mode {
    case Lookup:
        opt.DatFile = dem.lookup
        opt.ImgFmt = "FCOMPLEX"
        opt.Cmd = "mph"
    case Dem:
        opt.Datfile = dem.dat
        opt.Cmd = "hgt"
    default:
        return Handle(nil, "unrecognized plot mode")
    }
    
    return Raster(dem, opt)    
}


func NewGeocode(dat, par string) (ret Geocode, err error) {
    ret.MLI, err = NewMLI(dat, par)
    return
} 

func (opt *GeoPlotOpt) Parse() error {
    err := opt.rasArgs.Parse()
    
    if opt.cycle == 0 {
        opt.cycle = default_value
    }
    
    return nil
}

var rashgt = Gamma.must("rashgt")

// TODO: implement
func (geo *Geocode) Raster(opt GeoPlotOpt) error {
    err := opt.Parse(geo)
    
    return rashgt(geo.hgt, geo.MLI.dat, opt.Rng, opt.startHgt, opt.startPwr,
                  opt.Nlines, opt.Avg.Rng, opt.Avg.Azi, opt.cycle, opt.Scale,
                  opt.Exp, opt.LR, raster)
}

    

func (geo* Geocoder) Run(opt GeocodeOpt) (ret Geocode, err error) {
    out := geo.OutDir
    
    demdir := fp.Join(out, "dem")
    geodir := fp.Join(out, "geo")
    
    demOrig, err := NewDEM(fp.Join(demdir, "srtm.dem"), "")
    
    vertPath := geo.demPath
    
    if len(vrtPath) == 0 {
        err = Handle()
        return
    }
    
    ovsamp := geo.DEMOverSampling
    
    if ovsamp.Lat == 0.0 {
        ovsamp.Lat = 2.0
    }
    
    if ovsamp.Lon == 0.0 {
        ovsamp.Lon = 2.0
    }

    overlap := geo.DEMOverlap
    
    if overlap.Rng == 0 {
        overlap.Rng = 100
    }
    
    if overlap.Azi == 0 {
        overlap.Azi = 100
    }

    n_rng_off = params.getint("n_rng_off", 64)
    n_azi_off = params.getint("n_azi_off", 32)


    npoly, itr := geo.NPoly, geo.Iter
    
    if npoly == 0 {
        npoly = 4
    }
    
    
    ex, err := Exist(demOrig.dat)
    
    if err != nil {
        err = Handle()
        return
    }
    
    if !ex {
        log.Printf("Creating DEM from %s\n", vrtPath)
        
        vrt2dem(vrtPath, geo.mli.par, demOrig.dat, demOrig.par, 2, "-")
    } else {
        log.Println("DEM already imported.")
    }
            
    
    mra := RngAzi{}
    
    mra.Rng, err = geo.mli.Rng()
    
    if err != nil {
        err = Handle()
        return
    }
    
    mra.Azi, err = geo.mli.Azi()
    
    if err != nil {
        err = Handle()
        return
    }
    
    Patch := RngAzi{
        Rng: int(float64(mra.Rng) / off.Rng + float64(overlap.Rng) / 2),
        Azi: int(float64(mra.Azi) / off.Azi + float64(overlap.Azi) / 2),
    }
    
    // make sure the number of patches are even
    
    if Patch.Rng % 2 {
        Patch.Rng += 1
    }
    
    if Patch.Azi % 2 {
        Patch.Azi += 1
    }
    
    dem, err := NewDEM(fp.Join(dempath, "dem_seg.dem")
    
    dem = DEM(djoin("dem_seg.dem"), parfile=djoin("dem_seg.dem_par"),
              lookup=gjoin("lookup"), lookup_old=gjoin("lookup_old"))
    
    
    geo, err := NewGeocode(geo.mli.dat, geo.mli.par)
    
    if err != nil {
        err = Handle()
        return
    }
    
    geo := Geocode{
        MLI.dat : geo.mli.dat,
        MLI.par : geo.mli.par,
        hgt     : fp.Join(geopath, "hgt"),
        simSar  : fp.Join(geopath, "sim_sar"),
        zenith  : fp.Join(geopath, "zenith"),
        orient  : fp.Join(geopath, "orient"),
        orient  : fp.Join(geopath, "orient"),
        inc     : fp.Join(geopath, "inc"),
        pix     : fp.Join(geopath, "pix"),
        psi     : fp.Join(geopath, "psi"),
        lsMap   : fp.Join(geopath, "lsMap"),
        diffPar : fp.Join(geopath, "diff_par"),
        offs    : fp.Join(geopath, "offs"),
        offsets : fp.Join(geopath, "offsets"),
        ccp     : fp.Join(geopath, "ccp"),
        coffs   : fp.Join(geopath, "coffs"),
        coffsets: fp.Join(geopath, "coffsets"),
    }
    
    ex1, err := Exist(dem.lookup)
    
    ex2, err := Exist(dem.par)
    
    
    if !ex1 and !ex2 {
        log.Println("Calculating initial lookup table.")
        
        err = gcMap(geo.mli.par, "-", demOrig.par, demOrig.dat,
                    dem.par, dem.dat, dem.lookup, oversamp.Lat, oversamp.Lon,
                    geo.simSar, geo.zenith, geo.orient, geo.inc, geo.psi,
                    geo.pix, geo.lsMap, 8, 2)
        
        if err != nil {
            err = Handle()
            return
        }      
    } else {
        log.Println("Initial lookup table already created.")
    }
    
    dra := RngAzi{}
    
    dra.Rng, err = dem.Rng()
    
    if err != nil {
        err = Handle()
        return
    }
    
    dra.Azi, err = dem.Azi()
    
    if err != nil {
        err = Handle()
        return
    }
    
    // TODO: 20 magic number
    err = pixelArea(geo.mli.par, dem.par, dem.dat, dem.lookup, geo.lsMap,
                    geo.inc, geo.sigma0, geo.gamma0, 20)
    
    if err != nil {
        err = Handle()
        return
    }
    
    // TODO: magic numbers, 1, 0
    err = createDiffPar(geo.mli.par, "-", geo.diffPar, 1, 0)
    
    if err != nil {
        err = Handle()
        return
    }
    
    log.Println("Refining lookup table.")

    if itr >= 1 {
        log.Println("ITERATING OFFSET REFINEMENT.")
        
        for ii := 0; ii < itr; ii++ {
            log.Printf("ITERATION %d / %d\n", ii + 1, itr)
            
            rm(geo.diffPar)

            // copy previous lookup table
            cp(dem.lookup, dem.lookup_old)

            gp.create_diff_par(m_mli.par, None, geo.diff_par, 1, 0)

            gp.offset_pwrm(geo.sigma0, m_mli.dat, geo.diff_par, geo.offs,
                           geo.ccp, rng_patch, azi_patch, geo.offsets, 2,
                           n_rng_off, n_azi_off, 0.1, 5, 0.8)

            gp.offset_fitm(geo.offs, geo.ccp, geo.diff_par, geo.coffs,
                           geo.coffsets, 0.1, npoly)

            # update previous lookup table
            gp.gc_map_fine(dem.lookup_old, dem_s_width, geo.diff_par,
                           dem.lookup, 1)

            # create new simulated ampliutides with the new lookup table
            gp.pixel_area(m_mli.par, dem.par, dem.dat, dem.lookup, geo.ls_map,
                          geo.inc, geo.sigma0, geo.gamma0, 20)

        # end for
        log.info("ITERATION DONE.")
    # end if
    
    
    return {
        "geo": geo,
        "dem_orig": dem_orig,
        "dem": dem
    }


