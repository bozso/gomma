package geo

import (
    "log"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
    "github.com/bozso/gamma/base"
)


type Hgt struct {
    data.File
}

func (h *Hgt) Set(s string) (err error) {
    return
}

func (h Hgt) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = Height
    opt.Parse(h)
    
    err = plot.Raster(h, opt)
    return nil
}

type Geocode struct {
    MLI base.MLI
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
    createDiffPar = common.Gamma.Must("create_diff_par")
    vrt2dem = common.Gamma.Must("vrt2dem")
    //gcMap = common.Gamma.Must("gc_map2")
    gcMap = common.Gamma.Must("gc_map")
    pixelArea = common.Gamma.Must("pixel_area")
    offsetPwrm = common.Gamma.Must("offset_pwrm")
    offsetFitm = common.Gamma.Must("offset_fitm")
    gcMapFine = common.Gamma.Must("gc_map_fine")
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

