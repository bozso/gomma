package service

import (
    "fmt"
    "log"
    "io"
    "bufio"
    "net/http"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/emath/geometry"
    
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/common"
    s1 "github.com/bozso/gomma/sentinel1"
)

type DataSelect struct {
    CacheDir path.Dir
}

func (ds *DataSelect) Default() {
    if len(ds.CacheDir.String()) == 0 {
        ds.CacheDir, _ = path.New(".").ToDir()
    }
}

type SentinelSelect struct {
    Output
    DataFiles []path.ValidFile  `json:"data_files"`
    Start     date.ShortTime    `json:"start"`
    Stop      date.ShortTime    `json:"stop"`
    Region    geometry.Region   `json:"region"`
    AOI       common.AOI        `json:"aoi"`
    CheckZips bool              `json:"check_zips"`
    Pol       common.Pol        `json:"polarization"`
}

func (ds *DataSelect) SelectFiles(_ http.Request, ss *SentinelSelect, _ *Empty) (err error) {
    dataFiles := ss.DataFiles

    if len(dataFiles) == 0 {
        return fmt.Errorf("At least one datafile must be specified!")
    }
    
    checker := date.NewCheckers()

    if d := ss.Start; d.IsSet() {
        checker.Append(date.Min.New(d.Time))
    }
    if d := ss.Stop; d.IsSet() {
        checker.Append(date.Max.New(d.Time))
    }
    
    // TODO: implement checkZip
    //if dsect.CheckZips {
    //    checker = func(s1zip S1Zip) bool {
    //        return checker(s1zip) && s1zip.checkZip()
    //    }
    //    check = true
    //
    //}
    
    // nzip := len(zipfiles)
    
    
    writer := ss.Out
    defer writer.Close()
    
    for _, zip := range dataFiles {
        s1zip, IWs, err := parseS1(zip, ds.CacheDir)
        
        if err != nil {
            return err
        }
        
        if IWs.Contains(ss.AOI) && checker.In(s1zip.Date()) {
            _, err := fmt.Fprintf(writer, "%s\n", s1zip.Path)
            if err != nil {
                return err
            }
        }
    }
    
    return
}

func parseS1(zip path.ValidFile, dst path.Dir) (S1 *s1.Zip, IWs s1.IWInfos, err error) {
    if S1, err = s1.NewZip(zip); err != nil {
        return
    }
    
    log.Printf("Parsing IW Information for S1 zipfile '%s'", S1.Path)
    
    if IWs, err = S1.Info(dst); err != nil {
        return
    }
    
    return
}

func loadS1(reader io.Reader, pol string) (S1 s1.Zips, err error) {
    file := bufio.NewScanner(reader)
    
    for file.Scan() {
        if err = file.Err(); err != nil {
            return
        }

        vf, Err := path.New(file.Text()).ToValidFile()
        if Err != nil {
            err = Err
            return
        }
        
        s1zip, Err := s1.NewZip(vf)
        if Err != nil {
            err = Err
            return
        }
        
        S1 = append(S1, s1zip)
    }
    
    return
}


//func (s *selector) extOpt(satellite string) *ExtractOpt {
    //return &ExtractOpt{pol: s.Pol, 
        //root: filepath.Join(s.CachePath, satellite)}
//}


/*
type dataImporter struct {
    GeneralOpt
    IWs        [3]IMinmax `cli:"iws" usage:"IW burst indices"`
    
}

var DataImport = &cli.Command{
    Name: "import",
    Desc: "Import SAR datafiles",
    Argv: func() interface{} { return &dataImporter{} },
    Fn: dataImport,
}

var s1Import = Gamma.Must("S1_import_SLC_from_zipfiles")

func dataImport(ctx *cli.Context) (err error) {
    const (
        tpl = "iw%d_number_of_bursts: %d\niw%d_first_burst: %f\niw%d_last_burst: %f\n"
        burst_table = "burst_number_table"
        ziplist = "ziplist"
    )
    
    imp := ctx.Argv().(*dataImporter)
        
    defer imp.InFile.Close()
    var zips S1Zips
    if zips, err = loadS1(imp.InFile, imp.Pol); err != nil {
        return Handle(err, "failed to load zipfiles")
    }
    
    var master *S1Zip
    
    //sort.Sort(ByDate(zips))
    
    masterDate := imp.MasterDate
    
    for _, s1zip := range zips {
        if date2str(s1zip, DShort) == masterDate {
            master = s1zip
        }
    }
    
    
    var masterIW IWInfos
    if masterIW, err = master.Info(imp.CachePath); err != nil {
        return Handle(err, "failed to parse S1Zip data from master '%s'",
            master.Path)
    }
    
    fburst := NewWriterFile(burst_table);
    if err = fburst.Wrap(); err != nil {
        return
    }
    defer fburst.Close()
    
    fburst.WriteFmt("zipfile: %s\n", master.Path)
    
    nIWs := 0
    
    for ii, iw := range imp.IWs {
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
        
        fburst.WriteString(line)
        nIWs++
    }
    
    if err = fburst.Wrap(); err != nil {
        return
    }
    
    // defer os.Remove(ziplist)
    
    slcDir := filepath.Join(imp.OutputDir, "SLC")

    if err = os.MkdirAll(slcDir, os.ModePerm); err != nil {
        return DirCreateErr.Wrap(err, slcDir)
    }
    
    pol, writer := imp.Pol, bufio.NewWriter(&imp.OutFile)
    defer imp.OutFile.Close()
    
    for _, s1zip := range zips {
        // iw, err := s1zip.Info(extInfo)
        
        date := date2str(s1zip, DShort)
        
        other := Search(s1zip, zips)
        
        err = toZiplist(ziplist, s1zip, other)
        
        if err != nil {
            return Handle(err, "could not write zipfiles to zipfile list file '%s'",
                ziplist)
        }
        
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
            return
        }
        
        if _, err = writer.WriteString(slc.Tab); err != nil {
            return
        }
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

func stepGeocode (c *Config) error {
    outDir := c.General.OutputDir
    
    if err := c.Geocoding.Run(outDir); err != nil {
        return Handle(err, "geocoding failed")
    }
        
    return nil
}

func stepCoreg(self *Config) (err error) {
    outDir := self.General.OutputDir
    coreg := self.Coreg
    
    path := self.InFile
    
    var file FileReader
    if file, err = NewReader(path); err != nil {
        err = FileOpenErr.Make(path)
        return
    }
    defer file.Close()
    
    S1SLCs := []S1SLC{}
    
    for file.Scan() {
        line := strings.TrimSpace(file.Text())
        
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
        return StructCreateError.Make("MLI")
    }

    var hgt Hgt
    if err = Load(coreg.Hgt, &hgt); err != nil {
        return StructCreateError.Make("HGT")
    }
    
    rslc, ifg := filepath.Join(outDir, "RSLC"), filepath.Join(outDir, "IFG")
    
    if err = os.MkdirAll(rslc, os.ModePerm); err != nil {
        return DirCreateErr.Make(rslc)
    }
    
    if err = os.MkdirAll(ifg, os.ModePerm); err != nil {
        return DirCreateErr.Make(ifg)
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
        
        if err = out.Ifg.Raster(opt); err != nil {
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
        
        if err = out.Ifg.Raster(opt); err != nil {
            return Handle(err, "failed to create raster image for interferogram '%s",
                out.Ifg.Datfile())
        }
        
        prev = &out.Rslc
    }
    
    return nil
}
*/

func Search(s1 *s1.Zip, zips s1.Zips) *s1.Zip {
    date1 := date.Short.Format(s1)
    
    for _, zip := range zips {
        if date1 == date.Short.Format(zip)  && s1.Path != zip.Path {
            return zip
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
