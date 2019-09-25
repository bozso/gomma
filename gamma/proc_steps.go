package gamma


import (
    "fmt"
    "log"
    "sort"
    "math"
    //"time"
    fp "path/filepath"
    //conv "strconv"
    //str "strings"
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


func stepImport(self *config) error {
    if len(self.infile) == 0 {
        return Handle(nil, "inputfile must by specified")
    }
    
    meta := Meta{}
    path := self.General.Metafile
    err := LoadJson(path, &meta)
    
    if err != nil {
        return Handle(err, "failed to parse meta json file '%s'", path)
    }
    
    extInfo := self.extOpt("sentinel1")
    
    root := extInfo.root
    path = self.infile
    
    zips, err := loadS1(path, root)
    
    if err != nil {
        return Handle(err, "failed to load zipfiles from '%s'", path)
    }
    
    var (
        master *S1Zip
        idx int
    )
    
    sort.Sort(ByDate(zips))
    
    masterDate := meta.MasterDate
    
    for ii, s1zip := range zips {
        if date2str(s1zip, short) == masterDate {
            master = s1zip
            idx = ii
        }
    }
    
    meta.MasterIdx = idx
    
    masterIW, err := master.Info(extInfo)
    if err != nil {
        return Handle(err, "failed to parse S1Zip data from master '%s'",
            master.Path)
    }
    
    for _, s1zip := range zips {
        iw, err := s1zip.Info(extInfo)
        
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
        
    }
    
    path = self.General.Metafile
    err = SaveJson(path, meta)
    
    if err != nil {
        return Handle(err, "failed to write metadata to '%s'", path)
    }
    
    return nil
}


func stepGeocode (c *config) error {
    mzip := c.General.MasterImage
    
    root := fp.Join(c.General.CachePath, "sentinel1")
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
    err = SaveJson(fp.Join(outDir, "geocode.json"), &geocode)
    
    if err != nil {
        return Handle(err, "failed to save geocode data")
    }
    
    return nil
}

func stepCoreg(self *config) error {
    path, pol := self.General.Metafile, self.General.Pol
    
    root, meta := fp.Join(self.General.CachePath, "sentinel1"), Meta{}
    err := LoadJson(path, &meta)
    
    if err != nil {
        return Handle(err, "failed to read metadata from '%s'", path)
    }
    
    midx := meta.MasterIdx
    
    path = self.infile
    file, err := NewReader(path)
    
    if err != nil {
        return Handle(err, "failed to open file '%s'", path)
    }
    
    defer file.Close()
    
    s1zips := S1Zips{}
    
    for file.Scan() {
        line := file.Text()
        s1, err := NewS1Zip(line, root)
        
        if err != nil {
            return Handle(err, "failed to parse S1Zip data from '%s'",
                line)
        }
        
        s1zips = append(s1zips, s1)
    }
    
    mzip := s1zips[midx]
    
    fmt.Printf("Master date: %s\n", mzip.Center())
    
    mslc, err := mzip.SLC(pol)
    
    if err != nil {
        return Handle(err, "failed to import master SLC data")
    }
    
    master := S1Coreg{
        pol: pol,
        tab: mslc.tab,
        ID: date2str(mzip, short),
        coreg: self.Coreg,
        hgt: "0.1",
        poly1: "-",
        poly2: "-",
        Looks: self.General.Looks,
        clean: false,
        useInter: true, 
    }
    
    var prev *S1Zip
    nzip := len(s1zips)
    
    for ii := midx + 1; ii < nzip; ii++ {
        curr := s1zips[ii]
        
        ok, err := master.Coreg(curr, prev)
        
        if err != nil {
            return Handle(err, "coregistration failed")
        }
        
        fmt.Printf("%s\n", err)
        
        if !ok {
            log.Printf("%s\n",
                Handle(err, "coregistration of '%s' failed", curr.Path))
            log.Fatalf("First.\n")
            continue
        }
        
        log.Fatalf("First.\n")
        
        prev = curr
    }
    
    
    prev = nil
    
    for ii := midx - 1; ii > -1; ii-- {
        curr := s1zips[ii]
        
        ok, err := master.Coreg(curr, prev)
        
        if err != nil {
            return Handle(err, "coregistration failed")
        }
        
        if !ok {
            log.Printf("%s\n", Handle(err, "coregistration of '%s' failed!",
                curr.Center()))
            continue
        }
        
        prev = curr
    }
    
    /*
    
    for ii, S1 := range s1zips {
        if ii == midx {
            continue
        }
        
        date1 := S1.Center()
        date2 := s1zips[0].Center()
        
        idx, diff1 := 0, math.Abs(float64(date1.Sub(date2)))
        
        for jj := 1; jj < nzip; jj++ {
            diff2 := math.Abs(float64(date1.Sub(s1zips[jj].Center())))
            
            if diff2 < diff1 {
                idx = jj
            }
        }
        
        // S1Coreg(mslc, )
    }
    */
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
