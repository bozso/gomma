package gamma


import (
    "fmt"
    "log"
    "sort"
    "os"
    "path/filepath"
    // "math"
    // "time"
    //"strconv"
    //"strings"
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


func (self *Config) extOpt(satellite string) *ExtractOpt {
    return &ExtractOpt{pol: self.General.Pol, 
        root: filepath.Join(self.General.CachePath, satellite)}
}

func stepSelect(self *Config) error {
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

    zipfiles, err := filepath.Glob(filepath.Join(dataPath, "S1*_IW_SLC*.zip"))
    if err != nil {
        return Handle(err, "failed to Glob zipfiles")
    }
    
    var startCheck, stopCheck checkerFun
    
    checker := func(s1zip *S1Zip) bool {
        return true
    }
    
    if len(dateStart) != 0 {
        _dateStart, err := ParseDate(DShort, dateStart)
        
        if err != nil {
            return Handle(err, "failed to parse date '%s' in short format",
                dateStart)
        }
        
        startCheck = func(s1zip *S1Zip) bool {
            return s1zip.Start().After(_dateStart)
        }
    }
    
    if len(dateStop) != 0 {
        _dateStop, err := ParseDate(DShort, dateStop)
        
        if err != nil {
            return Handle(err, "failed to parse date '%s' in short format",
                dateStop)
        }
        
        stopCheck = func(s1zip *S1Zip) bool {
            return s1zip.Stop().Before(_dateStop)
        }
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
    
    
    var (
        s1zip *S1Zip
        IWs IWInfos
    )
    
    for _, zip := range zipfiles {
        if s1zip, IWs, err = parseS1(zip, root, extInfo); err != nil {
            return Handle(err,
                "failed to import S1Zip data from '%s'", zip)
        }
        
        if IWs.contains(aoi) && checker(s1zip) {
            fmt.Printf("%s\n", s1zip.Path)
        }
    }
    
    return nil
}

var s1Import = Gamma.Must("S1_import_SLC_from_zipfiles")

func stepImport(self *Config) error {
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
        if date2str(s1zip, DShort) == masterDate {
            master = s1zip
        }
    }
    
    masterIW, err := master.Info(extInfo)
    
    if err != nil {
        return Handle(err, "failed to parse S1Zip data from master '%s'",
            master.Path)
    }
    
    var fburst *os.File
    if fburst, err = os.Create(burst_table); err != nil {
        return FileOpenErr.Wrap(err, burst_table)
    }
    
    defer fburst.Close()
    
    _, err = fburst.WriteString(fmt.Sprintf("zipfile: %s\n", master.Path))
    if err != nil {
        return FileWriteErr.Wrap(err, burst_table)
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
        
        if _, err := fburst.WriteString(line); err != nil {
            return FileWriteErr.Wrap(err, burst_table)
        }
        nIWs++
    }
    
    // defer os.Remove(ziplist)
    
    slcDir := filepath.Join(self.General.OutputDir, "SLC")
    
    if err = os.MkdirAll(slcDir, os.ModePerm); err != nil {
        return DirCreateErr.Wrap(err, slcDir)
    }
    
    for _, s1zip := range zips {
        // iw, err := s1zip.Info(extInfo)
        
        date := date2str(s1zip, DShort)
        
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
            Tab: base + ".SLC_tab",
            nIW: nIWs,
        }
        
        
        for ii := 0; ii < nIWs; ii++ {
            dat := fmt.Sprintf("%s.slc.iw%d", base, ii + 1)
            
            slc.IWs[ii].Dat = dat
            slc.IWs[ii].Params = NewGammaParam(dat + ".par")
            slc.IWs[ii].TOPS_par = NewGammaParam(dat + ".TOPS_par")
        }
        
        if slc, err = slc.Move(slcDir); err != nil {
            return err
        }
        
        fmt.Println(slc.Tab)
    }
    
    // TODO: save master idx?
    //err = SaveJson(path, meta)
    //
    //if err != nil {
    //    return Handle(err, "failed to write metadata to '%s'", path)
    //}
    
    return nil
}

func stepGeocode (c *Config) error {
    outDir := c.General.OutputDir
    
    if err := c.Geocoding.Run(outDir); err != nil {
        return Handle(err, "geocoding failed")
    }
        
    return nil
}


func stepCoreg(self *Config) error {
    outDir := self.General.OutputDir
    coreg := self.Coreg
    
    path := self.infile
    file, err := NewReader(path)
    
    if err != nil {
        return Handle(err, "failed to open file '%s'", path)
    }
    
    defer file.Close()
    
    S1SLCs := []S1SLC{}
    
    for file.Scan() {
        line := str.TrimSpace(file.Text())
        
        var s1 S1SLC
        if s1, err = FromTabfile(line); err != nil {
            return Handle(err, "failed to parse S1SLC file from '%s'",
                line)
        }
        
        S1SLCs = append(S1SLCs, s1)
    }
    
    midx := self.Coreg.MasterIdx - 1
    
    mslc := S1SLCs[midx]
    mdate := mslc.Format(DateShort)
    
    fmt.Printf("Master date: %s\n", mdate)
    
    
    if len(coreg.Mli) == 0 || len(coreg.Hgt) == 0 {
        return fmt.Errorf("Path to master MLI file and path to " + 
                          "elevation model in radar coordinates " +
                          "has to be given!")
    }
    
    var mli MLI
    if err = Load(coreg.Mli, &mli); err != nil {
        return Handle(err, "failed to make master MLI struct")
    }

    var hgt Hgt
    if err = Load(coreg.Hgt, &hgt); err != nil {
        return Handle(err, "failed to make master MLI struct")
    }
    
    rslc, ifg := fp.Join(outDir, "RSLC"), fp.Join(outDir, "IFG")
    
    if err = os.MkdirAll(rslc, os.ModePerm); err != nil {
        return Handle(err, "failed to create directory '%s'", rslc)
    }
    
    if err = os.MkdirAll(ifg, os.ModePerm); err != nil {
        return Handle(err, "failed to create directory '%s'", ifg)
    }
    
    master := S1Coreg{
        Tab: mslc.Tab,
        ID: mdate,
        CoregOpt: self.Coreg,
        Hgt: hgt.Dat,
        Poly1: "-",
        Poly2: "-",
        Looks: self.General.Looks,
        Clean: false,
        UseInter: true,
        OutDir: outDir,
        RslcPath: rslc,
        IfgPath: ifg,
    }
    
    var prev *S1SLC = nil
    nzip := len(S1SLCs)
    
    opt := RasArgs{DisArgs:DisArgs{Sec: mli.Dat}}
    
    for ii := midx + 1; ii < nzip; ii++ {
        curr := &S1SLCs[ii]
        
        out, err := master.Coreg(curr, prev)
        
        if err != nil {
            return Handle(err, "coregistration failed")
        }
        
        if !out.Ok {
            log.Printf("Coregistration of '%s' failed! Moving to the next scene\n",
                curr.Format(DateShort))
            continue
        }
        
        err = out.Ifg.Raster(opt)
        
        if err != nil {
            return Handle(err, "failed to create raster image for interferogram '%s",
                out.Ifg.Dat)
        }
        
        prev = &out.Rslc
    }
    
    
    prev = nil
    
    for ii := midx - 1; ii > -1; ii-- {
        curr := &S1SLCs[ii]
        
        out, err := master.Coreg(curr, prev)
        
        if err != nil {
            return Handle(err, "coregistration failed")
        }
        
        if !out.Ok {
            log.Printf("Coregistration of '%s' failed! Moving to the next scene\n",
                curr.Format(DateShort))
            continue
        }
        
        err = out.Ifg.Raster(opt)
        
        if err != nil {
            return Handle(err, "failed to create raster image for interferogram '%s",
                out.Ifg.Datfile())
        }
        
        prev = &out.Rslc
    }
    
    return nil
}


func Search(s1 *S1Zip, zips []*S1Zip) *S1Zip {
    
    date1 := date2str(s1, DShort)
    
    for _, zip := range zips {
        if date1 == date2str(zip, DShort)  && s1.Path != zip.Path {
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
        if _, err = file.WriteString(one.Path); err != nil {
            return Handle(err, "failed to write to ziplist file '%s'", name)
        }
    } else {
        after := two.date.center.After(one.date.center)
        
        if after {
            if _, err = file.WriteString(one.Path + "\n"); err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
            
            if _, err = file.WriteString(two.Path + "\n"); err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
        } else {
            if _, err = file.WriteString(two.Path + "\n"); err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
            
            if _, err = file.WriteString(one.Path + "\n"); err != nil {
                return Handle(err, "failed to write to ziplist file '%s'", name)
            }
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
