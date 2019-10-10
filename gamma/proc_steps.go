package gamma


import (
    "fmt"
    "log"
    "sort"
    "os"
    // "math"
    // "time"
    fp "path/filepath"
    //conv "strconv"
    // str "strings"
)

type(
    SARInfo interface {
        contains(*AOI) bool
    }
    
    
    SARImage interface {
        Info(*ExtractOpt) (SARInfo, error)
        SLC(*ExtractOpt) (SLC, error)
    }
    
    checkerFun func(*S1Zip) bool
)


func parseS1(zip, root string, ext *ExtractOpt) (s1 *S1Zip, IWs IWInfos, err error) {
    s1, err = NewS1Zip(zip, root)
    
    if err != nil {
        err = Handle(err, "failed to parse S1Zip data from '%s'", zip)
        return
    }

    log.Printf("Parsing IW Information for S1 zipfile '%s'", s1.Path)
    
    IWs, err = s1.Info(ext)
    
    if err != nil {
        err = Handle(err, "failed to parse IW information for zip '%s'",
            s1.Path)
        return
    }
    
    return s1, IWs, nil
}

func loadS1(path, root string) (ret S1Zips, err error) {
    file, err := NewReader(path)
    
    if err != nil {
        err = Handle(err, "failed to open file '%s'", path)
        return
    }
    
    defer file.Close()
    
    var s1zip *S1Zip
    
    for file.Scan() {
        line := file.Text()
        s1zip, err = NewS1Zip(line, root)
        
        if err != nil {
            err = Handle(err, "failed to parse zipfile '%s'", line)
            return
        }
        
        ret = append(ret, s1zip)
    }
    
    return ret, nil
}


func (self *config) extOpt(satellite string) *ExtractOpt {
    return &ExtractOpt{pol: self.General.Pol, 
        root: fp.Join(self.General.CachePath, satellite)}
}

func stepSelect(self *config) error {
    dataPath := self.General.DataPath
    Select := self.PreSelect
    
    if len(dataPath) == 0 {
        return fmt.Errorf("DataPath needs to be specified")
    }

    ll, ur := Select.LowerLeft, Select.UpperRight

    aoi := AOI{
        Point{X: ll.Lon, Y: ll.Lat}, Point{X: ll.Lon, Y: ur.Lat},
        Point{X: ur.Lon, Y: ur.Lat}, Point{X: ur.Lon, Y: ll.Lat},
    }
    
    extInfo := self.extOpt("sentinel1")
    root := extInfo.root
    
    dateStart, dateStop := Select.DateStart, Select.DateStop

    zipfiles, err := fp.Glob(fp.Join(dataPath, "S1*_IW_SLC*.zip"))
    if err != nil {
        return Handle(err, "failed to Glob zipfiles")
    }
    
    
    var checker, startCheck, stopCheck checkerFun
    check := false
    
    
    if len(dateStart) != 0 {
        _dateStart, err := ParseDate(short, dateStart)
        
        if err != nil {
            return Handle(err, "failed to parse date '%s' in short format",
                dateStart)
        }
        
        startCheck = func(s1zip *S1Zip) bool {
            return s1zip.Start().After(_dateStart)
        }
        check = true
    }
    
    if len(dateStop) != 0 {
        _dateStop, err := ParseDate(short, dateStop)
        
        if err != nil {
            return Handle(err, "failed to parse date '%s' in short format",
                dateStop)
        }
        
        stopCheck = func(s1zip *S1Zip) bool {
            return s1zip.Stop().Before(_dateStop)
        }
        check = true
    }
    
    if startCheck != nil && stopCheck != nil {
        checker = func(s1zip *S1Zip) bool {
            return startCheck(s1zip) && stopCheck(s1zip)
        }
    } else if startCheck != nil {
        checker = startCheck
    } else if stopCheck != nil {
        checker = stopCheck
    }
    
    
    // TODO: implement checkZip
    //if Select.CheckZips {
    //    checker = func(s1zip S1Zip) bool {
    //        return checker(s1zip) && s1zip.checkZip()
    //    }
    //    check = true
    //
    //}
    
    // nzip := len(zipfiles)
    
    
    
    if check {
        for _, zip := range zipfiles {
            s1zip, IWs, err := parseS1(zip, root, extInfo)
            
            if err != nil {
                return Handle(err,
                    "failed to import S1Zip data from '%s'", zip)
            }
            
            if IWs.contains(aoi) && checker(s1zip) {
                fmt.Printf("%s\n", s1zip.Path)
            }
        }
    } else {
        for _, zip := range zipfiles {
            s1zip, IWs, err := parseS1(zip, root, extInfo)
            if err != nil {
                return Handle(err,
                    "failed to import S1Zip data from '%s'", zip)
            }
            
            if IWs.contains(aoi) {
                fmt.Printf("%s\n", s1zip.Path)
            }
        }
    }
    
    return nil
}

var s1Import = Gamma.must("S1_import_SLC_from_zipfiles")

func stepImport(self *config) error {
    const (
        tpl = "iw%d_number_of_bursts: %d\niw%d_first_burst: %f\niw%d_last_burst: %f\n"
        burst_table = "burst_number_table"
        ziplist = "ziplist"
    )
    
    pol := self.General.Pol
    
    if len(self.infile) == 0 {
        return Handle(nil, "inputfile must by specified")
    }
    
    extInfo := self.extOpt("sentinel1")
    
    root := extInfo.root
    path := self.infile
    
    zips, err := loadS1(path, root)
    
    if err != nil {
        return Handle(err, "failed to load zipfiles from '%s'", path)
    }
    
    var master *S1Zip = nil
    
    sort.Sort(ByDate(zips))
    
    masterDate := self.General.MasterDate
    
    for _, s1zip := range zips {
        if date2str(s1zip, short) == masterDate {
            master = s1zip
        }
    }
    
    masterIW, err := master.Info(extInfo)
    
    if err != nil {
        return Handle(err, "failed to parse S1Zip data from master '%s'",
            master.Path)
    }
    
    fburst, err := os.Create(burst_table)
    
    if err != nil {
        return Handle(err, "failed to open temporary file '%s'", burst_table)
    }
    
    defer fburst.Close()
    // defer os.Remove(burst_table)
    
    _, err = fburst.WriteString(fmt.Sprintf("zipfile: %s\n", master.Path))
    
    if err != nil {
        return Handle(err, "failed to write burst_number_table '%s'", burst_table)
    }
    
    nIWs := 0
    
    for ii, iw := range self.General.IWs {
        min, max := iw.Min, iw.Max
        
        if min == 0 && max == 0 {
            continue
        }
        
        nburst := max - min
        
        if nburst < 0 {
            return Handle(nil, "number of burst for IW%d is negative, did " +
                "you mix up first and last burst numbers?", ii + 1)
        }
        
        IW := masterIW[ii]
        
        first := IW.bursts[min - 1]
        last := IW.bursts[max - 1]
        
        line := fmt.Sprintf(tpl, ii + 1, nburst, ii + 1, first, ii + 1, last)
        
        _, err := fburst.WriteString(line)
        
        if err != nil {
            return Handle(err, "failed to write burst_number_table '%s'", burst_table)
        }
        nIWs++
    }
    
    // defer os.Remove(ziplist)
    
    slcDir := fp.Join(self.General.OutputDir, "SLC")
    
    err = os.MkdirAll(slcDir, os.ModePerm)
    
    if err != nil {
        return Handle(err, "failed to create directory '%s'", slcDir)
    }
    
    for _, s1zip := range zips {
        // iw, err := s1zip.Info(extInfo)
        
        date := date2str(s1zip, short)
        
        other := Search(s1zip, zips)
        
        err = toZiplist(ziplist, s1zip, other)
        
        if err != nil {
            return Handle(err, "could not write zipfiles to zipfile list file '%s'",
                ziplist)
        }
        
        // 0 = FCOMPLEX, 0 = swath_flag - as listed in burst_number_table_ref
        // "." = OPOD dir, 1 = intermediate files are deleted
        // 1 = apply noise correction
        _, err = s1Import(ziplist, burst_table, pol, 0, 0, ".", 1, 1)
        
        if err != nil {
            return Handle(err, "failed to import zipfile '%s'", s1zip.Path)
        }
        
        base := fmt.Sprintf("%s.%s", date, pol)
        
        slc := S1SLC{
            tab: base + ".SLC_tab",
            nIW: nIWs,
        }
        
        
        for ii := 0; ii < nIWs; ii++ {
            dat := fmt.Sprintf("%s.slc.iw%d", base, ii + 1)
            
            slc.IWs[ii].dataFile.Dat = dat
            slc.IWs[ii].dataFile.Params = NewGammaParam(dat + ".par")
            slc.IWs[ii].TOPS_par = NewGammaParam(dat + ".TOPS_par")
        }
        
        err = slc.Move(slcDir)
        
        if err != nil {
            return Handle(err, "failed to move S1SLC")
        }
        
        fmt.Println(slc.tab)
                
        /*
        if err != nil {
            return Handle(err, "failed to parse S1Zip data from '%s'",
                s1zip.Path)
        }
        
        if checkBurstNum(masterIW, iw) {
            log.Printf("S1Zip '%s' does not have the same number of " + 
                "bursts in every IW as the master image.", s1zip.Path)
            continue
        }
        
        diff, err := IWAbsDiff(masterIW, iw)
        
        if err != nil {
            return Handle(err,
            "failed to calculate burst number differences between " +
            "master and '%s'", s1zip.Path)
        }
        
        if !(math.RoundToEven(diff) > 0.0) {
            err = s1zip.ImportSLC(extInfo)
            
            if err != nil {
                return Handle(err, "failed to import S1SLC files")
            }
            
            fmt.Printf("%s\n", s1zip.Path)
        }
        */
    }
    
    // TODO: save master idx?
    //err = SaveJson(path, meta)
    //
    //if err != nil {
    //    return Handle(err, "failed to write metadata to '%s'", path)
    //}
    
    return nil
}


/*
func stepGeocode (c *config) error {
    mzip := c.General.MasterDate
    
    path, pol, outDir := c.infile, c.General.Pol, c.General.OutputDir
    
    zips, err := loadS1(path, root)
    
    if err != nil {
        return Handle(err, "failed to load zipfiles from '%s'", path)
    }
    
    var master *S1Zip
    
    if len(mzip) == 0 {
        sort.Sort(ByDate(zips))
        
        master = zips[0]
    } else {
        master, err = NewS1Zip(mzip, root)
        
        if err != nil {
            return Handle(err, "failed to parse master zipfile '%s'", mzip)
        }
    }
    
    mslc, err := master.SLC(pol)
    
    if err != nil {
        return Handle(err, "failed to retreive master S1SLC")
    }
    
    opt := &MLIOpt{
        refTab: mslc.tab,
        Looks: c.General.Looks,
    }
    
    mmli, err := master.MLI("mli", pol, opt)
    
    if err != nil {
        return Handle(err, "failed to retreive master MLI")
    }
    
    geo := Geocoder{
        geocoding: c.Geocoding,
        mli: mmli,
        outDir: outDir,
    }
    
    geocode, err := geo.Run()
    
    if err != nil {
        return Handle(err, "geocoding failed")
    }
    
    // TODO: make geocode.json a setting in configuration file?
    err = SaveJson(fp.Join(outDir, "geocode.json"), geocode)
    
    if err != nil {
        return Handle(err, "failed to save geocode data")
    }
    
    return nil
}


func stepCheckGeo(c *config) error {
    meta := GeoMeta{}
    path := fp.Join(c.General.OutputDir, "geocode.json")
    err := LoadJson(path, &meta)
    
    if err != nil {
        return Handle(err, "failed to parse meta json file '%s'", path)
    }
    
    geo, dem := meta.Geo, meta.Dem
    geo.MLI.sep = ":"
    dem.sep = ":"
    
    mrng, err := geo.MLI.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive master MLI range samples")
    }
    
    mazi, err := geo.MLI.Azi()
    
    if err != nil {
        return Handle(err, "failed to retreive master MLI azimuth lines")
    }
    
    drng, err := dem.Rng()
    
    if err != nil {
        return Handle(err, "failed to retreive DEM range samples")
    }
    

    log.Printf("Geocoding DEM heights into image coordinates.\n")
    
    
    opt := CodeOpt{
        inWidth: drng,
        outWidth: mrng,
        nlines: mazi,
        dtype: "FLOAT",
        interpolMode: InvSquaredDist,
    }
    
    err = dem.geo2radar(dem.Dat, geo.Hgt, opt)
    
    if err != nil {
        return Handle(err,
            "failed to geocode from geographic to radar coordinates")
    }
    
    popt := rasArgs{}
    
    err = dem.Raster(Lookup, popt)
    
    if err != nil {
        return Handle(err, "raster generation for DEM failed")
    }
    
    // TODO: make gm.raster2
    log.Printf("Creating quicklook hgt file.\n")
    
    popt2 := GeoPlotOpt{
        rasArgs: popt,
        cycle: 500.0,
    }
    
    err = geo.Raster(popt2)
    
    if err != nil {
        return Handle(err, "raster generation for HGT file failed")
    }
    
    // geo.raster("gamma0")

    // gp.dis2pwr(hgt.mli.dat, geo.gamma0, mrng, mrng)
    
    
    return nil
}
*/

//func stepCoreg(self *config) error {
    //mzip, pol := self.General.MasterImage, self.General.Pol
    
    //root := fp.Join(self.General.CachePath, "sentinel1")
    
    //path := self.infile
    //file, err := NewReader(path)
    
    //if err != nil {
        //return Handle(err, "failed to open file '%s'", path)
    //}
    
    //defer file.Close()
    
    //s1zips := S1Zips{}
    
    //ms1, err := NewS1Zip(mzip, root)
    
    //if err != nil {
        //return Handle(err, "failed to parse S1Zip data from '%s'",
            //mzip)
    //}
    
    //s1zips = append(s1zips, ms1)
    
    //for file.Scan() {
        //line := str.TrimSpace(file.Text())
        
        //if line == mzip {
            //continue
        //}
        
        //s1, err := NewS1Zip(line, root)
        
        //if err != nil {
            //return Handle(err, "failed to parse S1Zip data from '%s'",
                //line)
        //}
        
        //s1zips = append(s1zips, s1)
    //}
    
    //fmt.Printf("Master date: %s\n", ms1.Center())
    
    //mslc, err := ms1.SLC(pol)
    
    //if err != nil {
        //return Handle(err, "failed to import master SLC data")
    //}
    
    //meta := GeoMeta{}
    //path = fp.Join(self.General.OutputDir, "geocode.json")
    //err = LoadJson(path, &meta)
    
    //if err != nil {
        //return Handle(err, "failed to parse meta json file '%s'", path)
    //}
    
    //master := S1Coreg{
        //pol: pol,
        //tab: mslc.tab,
        //ID: date2str(ms1, short),
        //coreg: self.Coreg,
        //hgt: meta.Geo.Hgt,
        //poly1: "-",
        //poly2: "-",
        //Looks: self.General.Looks,
        //clean: false,
        //useInter: true, 
    //}
    
    //midx := 0
    
    //var prev *S1Zip
    //nzip := len(s1zips)
    
    //for ii := midx + 1; ii < nzip; ii++ {
        //curr := s1zips[ii]
        
        //ok, err := master.Coreg(curr, prev)
        
        //if err != nil {
            //return Handle(err, "coregistration failed")
        //}
        
        //if !ok {
            //log.Printf("Coregistration of '%s' failed! Moving to the next scene\n",
                //curr.Center())
            //continue
        //}
        
        //prev = curr
    //}
    
    
    //prev = nil
    
    //for ii := midx - 1; ii > -1; ii-- {
        //curr := s1zips[ii]
        
        //ok, err := master.Coreg(curr, prev)
        
        //if err != nil {
            //return Handle(err, "coregistration failed")
        //}
        
        //if !ok {
            //log.Printf("Coregistration of '%s' failed! Moving to the next scene\n",
                //curr.Center())
            //continue
        //}
        
        //prev = curr
    //}
    
    ///*
    
    //for ii, S1 := range s1zips {
        //if ii == midx {
            //continue
        //}
        
        //date1 := S1.Center()
        //date2 := s1zips[0].Center()
        
        //idx, diff1 := 0, math.Abs(float64(date1.Sub(date2)))
        
        //for jj := 1; jj < nzip; jj++ {
            //diff2 := math.Abs(float64(date1.Sub(s1zips[jj].Center())))
            
            //if diff2 < diff1 {
                //idx = jj
            //}
        //}
        
        //// S1Coreg(mslc, )
    //}
    //*/
    //return nil
//}


func Search(s1 *S1Zip, zips []*S1Zip) *S1Zip {
    
    date1 := date2str(s1, short)
    
    for _, zip := range zips {
        if date1 == date2str(zip, short)  && s1.Path != zip.Path {
            return zip
        }
    }
    
    return nil
}

func toZiplist(name string, one, two *S1Zip) error {
    file, err := os.Create(name)
    
    if err != nil {
        return Handle(err, "failed to open file '%s'", name)
    }
    
    defer file.Close()
    
    if two == nil {
        _, err = file.WriteString(one.Path)
        
        if err != nil {
            return Handle(err, "failed to write to ziplist file '%s'", name)
        }
    } else {
        after := two.date.center.After(one.date.center)
        
        if after {
            _, err = file.WriteString(one.Path + "\n")
            
            if err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
            
            _, err = file.WriteString(two.Path + "\n")
            
            if err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
        } else {
            _, err = file.WriteString(two.Path + "\n")
            
            if err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
            
            _, err = file.WriteString(one.Path + "\n")
            
            if err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
        }
    }
    return nil
}

func moveSLC(date, pol, slcDir string, nIWs int) error {
    base := fmt.Sprintf("%s.%s", date, pol)
    
    tab := fp.Join(slcDir, base + ".SLC_tab")
    
    file, err := os.Create(tab)
    
    if err != nil {
        return Handle(err, "could not open tabfile '%s'", tab)
    }
    
    defer file.Close()
    
    for ii := 0; ii < nIWs; ii++ {
        slc := fmt.Sprintf("%s.slc.iw%d", base, ii + 1)
        par := slc + ".par"
        TOPS_par := slc + ".TOPS_par"
        
        dst := fp.Join(slcDir, slc)
        err := os.Rename(slc, dst)

        if err != nil {
            return Handle(err, "failed to move file '%s' to '%s'", slc, dst)
        }
        
        slc = dst

        
        dst = fp.Join(slcDir, par)
        err = os.Rename(par, dst)

        if err != nil {
            return Handle(err, "failed to move file '%s' to '%s'", par, dst)
        }
        
        par = dst
        
        
        dst = fp.Join(slcDir, TOPS_par)
        err = os.Rename(TOPS_par, dst)

        if err != nil {
            return Handle(err, "failed to move file '%s' to '%s'", TOPS_par, dst)
        }
        
        TOPS_par = dst
        
        
        line := fmt.Sprintf("%s %s %s\n", slc, par, TOPS_par)
        
        _, err = file.WriteString(line)
        
        if err != nil {
            return Handle(err, "failed to write to tabfile '%s'", tab)
        }
    }
    
    return nil
}

/*
[check_ionosphere]
# range and azimuth window size used in offset estimation
rng_win = 256
azi_win = 256

# threshold value used in offset estimation
iono_thresh = 0.1

# range and azimuth step used in offset estimation,
# default (rng|azi)_win / 4
rng_step =
azi_step =


[reflector]
# station file containing reflector parameters
station_file = /mnt/Dszekcso/NET/D_160928.stn

# oversempling factor for SLC search
ref_ovs = 16

# size of search window
ref_win = 3
*/
