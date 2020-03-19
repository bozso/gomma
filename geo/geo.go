package geo

import (
    "log"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/dem"
    "github.com/bozso/gomma/base"
    "github.com/bozso/gomma/common"
)

type Geocode struct {
    DiffPar, Offs, Offsets, Ccp, Coffs, Coffsets path.File
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
    createDiffPar = common.Must("create_diff_par")
    vrt2dem = common.Must("vrt2dem")
    gcMap = common.Must("gc_map")
    pixelArea = common.Must("pixel_area")
    offsetPwrm = common.Must("offset_pwrm")
    offsetFitm = common.Must("offset_fitm")
    gcMapFine = common.Must("gc_map_fine")
    
    //gcMap = common.Must("gc_map2")
)

type GeocodeOpt struct {
    MasterMLI base.MLI
    DEMOverlap, OffsetWindows common.RngAzi
    Iter, NPoly, RngOversamp, nPixel, LanczosOrder int
    MLIOversamp int
    DEMOversamp common.LatLon
    VrtPath path.ValidFile
    CCThresh, AreaFactor, BandwithFrac float64
}

func (g* GeocodeOpt) Run(outDir path.Dir) (err error) {
    geodir, _ := outDir.Join("geo").ToDir()
    
    if err = geodir.Mkdir(); err != nil {
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
    
    demLoader := dem.New(geodir.Join("srtm.dem"))
    
    vrtPath := g.VrtPath

    ex, err := demLoader.DatFile.Exist()
    if err != nil {
        return
    }
    
    mli := g.MasterMLI
    
    if !ex {
        log.Printf("Creating DEM from %s\n", vrtPath)
        
        // magic number 2 = add interpolated geoid offset
        _, err = vrt2dem.Call(vrtPath, mli.ParFile,
            demLoader.DatFile, demLoader.ParFile, 2, "-")
        
        if err != nil {
            return
        }
    } else {
        log.Println("DEM already imported.")
    }

    originalDem, err := demLoader.Load()
    if err != nil {
        return
    }

    if err = originalDem.Save(); err != nil {
        return
    }
    
    mra := mli.Ra
    offsetWin := g.OffsetWindows
    
    Patch := common.RngAzi{
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
    
    demLoader = dem.New(geodir.Join("dem_seg.dem"))

    Geo := Geocode{
        Offs    : geodir.Join("offs").ToFile(),
        Offsets : geodir.Join("offsets").ToFile(),
        Ccp     : geodir.Join("ccp").ToFile(),
        Coffs   : geodir.Join("coffs").ToFile(),
        Coffsets: geodir.Join("coffsets").ToFile(),
        DiffPar : geodir.Join("diff_par").ToFile(),
    }
    
    sigma0 := data.New(geodir.Join("sigma0"))
    gamma0 := data.New(geodir.Join("gamma0"))
    lsMap := data.New(geodir.Join("lsmap"))
    simSar := data.New(geodir.Join("sim_sar"))
    zenith := data.New(geodir.Join("zenith"))
    orient := data.New(geodir.Join("orient"))
    inc := data.New(geodir.Join("inclination"))
    proj := data.New(geodir.Join("projection"))
    pix := data.New(geodir.Join("pixel_area"))

    lookup := dem.NewLookup(geodir.Join("lookup"))
    lookupOld := dem.NewLookup(geodir.Join("lookup_old"))
    
    ex1, err := lookup.DatFile.Exist()
    if err != nil {
        return
    }
    
    ex2, err := demLoader.ParFile.Exist()
    if err != nil {
        return
    }
    
    if !ex1 && !ex2 {
        log.Println("Calculating initial lookup table.")
        
        oversamp := g.DEMOversamp
        
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
        _, err = gcMap(mli.par, originalDem.par, originalDem.dat, dem.par, dem.dat,
                       dem.lookup, oversamp.Lat, oversamp.Lon, originalDem.lsMap,
                       geo.lsMap, originalDem.incidence, originalDem.resolution,
                       originalDem.offnadir, g.RngOversamp, Standard, NoMask,
                       g.nPixel, "-", Actual)
        */
        
        _, err = gcMap.Call(mli.ParFile, nil,
            originalDem.ParFile, originalDem.DatFile,
            demLoader.ParFile, demLoader.DatFile,
            lookup.DatFile, oversamp.Lat, oversamp.Lon,
            simSar.DatFile, zenith.DatFile, orient.DatFile,
            inc.DatFile, proj.DatFile, pix.DatFile,
            lsMap.DatFile, g.nPixel, 2, g.RngOversamp)
                
        if err != nil {
            return
        }      
    } else {
        log.Println("Initial lookup table already created.")
    }
    
    segmentedDem, err := demLoader.Load();
    if err != nil {
        return
    }
    
    dra := segmentedDem.Ra
    
    _, err = pixelArea.Call(mli.ParFile,
        segmentedDem.ParFile, segmentedDem.DatFile,
        lookup.DatFile, lsMap.DatFile,
        inc.DatFile, sigma0.DatFile, gamma0.DatFile, g.AreaFactor)
    
    if err != nil {
        return
    }
    
    _, err = createDiffPar.Call(mli.ParFile, nil, Geo.DiffPar, SLC_MLI, NonInter)
    
    if err != nil {
        return
    }
    
    log.Println("Refining lookup table.")
    
    if itr >= 1 {
        log.Println("ITERATING OFFSET REFINEMENT.")
        
        for ii := 0; ii < itr; ii++ {
            log.Printf("ITERATION %d / %d\n", ii + 1, itr)
            
            if err = Geo.DiffPar.Remove(); err != nil {
                return
            }

            // copy previous lookup table
            err = lookup.DatFile.Rename(lookupOld.DatFile)
            if err != nil {
                return
            }
            
            _, err = createDiffPar.Call(mli.ParFile, nil, Geo.DiffPar,
                                   SLC_MLI, NonInter)
            
            if err != nil { return }
            
            _, err = offsetPwrm.Call(sigma0.DatFile, mli.DatFile,
                Geo.DiffPar, Geo.Offs, Geo.Ccp, Patch.Rng, Patch.Azi,
                Geo.Offsets, g.MLIOversamp, offsetWin.Rng,
                offsetWin.Azi, g.CCThresh, g.LanczosOrder,
                g.BandwithFrac)
            
            if err != nil { return }
            
            _, err = offsetFitm.Call(Geo.Offs, Geo.Ccp, Geo.DiffPar,
                Geo.Coffs, Geo.Coffsets, g.CCThresh, npoly, NonInter)
            
            if err != nil { return }

            // update previous lookup table
            // TODO: magic number 1
            _, err = gcMapFine.Call(lookupOld.DatFile, dra.Rng,
                Geo.DiffPar, lookup.DatFile, 1)
            
            if err != nil { return }

            // create new simulated ampliutides with the new lookup table
            _, err = pixelArea.Call(mli.ParFile,
                segmentedDem.ParFile, segmentedDem.DatFile,
                lookup.DatFile, lsMap.DatFile, inc.DatFile,
                sigma0.DatFile, gamma0.DatFile, g.AreaFactor)
            
            if err != nil { return }

        }
        log.Println("ITERATION DONE.")
    }
    
    
    toLoad := []*data.Path{
        &sigma0, &gamma0, &lsMap, &simSar, &zenith, &orient, &inc, &proj, &pix,
    }
    
    toSave := make([]data.Saver, 0)
    
    for _, load := range toLoad {
        loaded, Err := mli.WithShape(*load)
        if err != nil {
            err = Err
            return
        }
        
        toSave = append(toSave, loaded)
    }
    
    lookupLoaded, err := segmentedDem.LoadLookup(lookup)
    if err != nil {
        return
    }
    
    toSave = append(toSave, &segmentedDem, &lookupLoaded)
    
    for _, s := range toSave {
        err = s.Save()
        
        if err != nil {
            return
        }
    }
    
    return nil
}

