package gamma


import (
    "fmt"
    "log"
    "io"
    "bufio"    
)

type checkerFun func(*S1Zip) bool

func parseS1(zip, pol, dst string) (s1 *S1Zip, IWs IWInfos, err error) {
    if s1, err = NewS1Zip(zip, pol); err != nil {
        err = Handle(err, "failed to parse S1Zip data from '%s'", zip)
        return
    }
    log.Printf("Parsing IW Information for S1 zipfile '%s'", s1.Path)
    
    if IWs, err = s1.Info(dst); err != nil {
        err = Handle(err, "failed to parse IW information for zip '%s'",
            s1.Path)
        return
    }
    
    return s1, IWs, nil
}

func loadS1(reader io.Reader, pol string) (s1 S1Zips, err error) {
    file := bufio.NewScanner(reader)
    
    for file.Scan() {
        line := file.Text()
        
        var s1zip *S1Zip
        if s1zip, err = NewS1Zip(line, pol); err != nil {
            err = Handle(err, "failed to parse zipfile '%s'", line)
            return
        }
        
        s1 = append(s1, s1zip)
    }
    
    return s1, nil
}


//func (s *selector) extOpt(satellite string) *ExtractOpt {
    //return &ExtractOpt{pol: s.Pol, 
        //root: filepath.Join(s.CachePath, satellite)}
//}

type GeneralOpt struct {
    //DataPath   string     `cli:"" usage:""`
    OutputDir, Pol, MasterDate, CachePath  string
    Looks      RngAzi
    InFile     Reader
    OutFile    Writer
}

func (g *GeneralOpt) SetCli(c *Cli) {
    g.InFile.SetCli(c, "infile", "Input file.")
    g.OutFile.SetCli(c, "outfile", "Input file.")
    
    c.StringVar(&g.OutputDir, "out", ".", "Output directory")
    c.StringVar(&g.Pol, "pol", "vv", "POlarisation used.")
    c.StringVar(&g.MasterDate, "masterDate", "", "")
    c.StringVar(&g.CachePath, "cachePath", "", "Cache path.")
    c.Var(&g.Looks, "looks", "Range, azimuth looks.")
}

type dataSelect struct {
    GeneralOpt
    DataFiles  Files `cli:"d,data" usage:"List of datafiles to process"`
    DateStart  string   `cli:"b,start" usage:"Start of date range"`
    DateStop   string   `cli:"e,stop" usage:"End of date range"`
    LowerLeft  LatLon   `cli:"ll,lowerLeft" usage:"Rectangle coordinates"`
    UpperRight LatLon   `cli:"ur,upperRight" usage:"Rectangle coordinates"`
    CheckZips  bool     `cli:"c,checkZips" usage:"Check zipfile integrity"`  
}

func (d *dataSelect) SetCli(c *Cli) {
    d.GeneralOpt.SetCli(c)
    
    c.Var(&d.DataFiles, "dataFiles",
        "Comma separated filpaths to Sentinel-1 data.")
    
    c.StringVar(&d.DateStart, "start", "", "Start of date range.")
    c.StringVar(&d.DateStop, "stop", "", "End of date range.")
    c.Var(&d.LowerLeft, "lowerLeft", "Rectangle coordinates.")
    c.Var(&d.UpperRight, "upperRight", "Rectangle coordinates.")
    c.BoolVar(&d.CheckZips, "checkZips", false, "Check zipfile integrity.")
}

func (sel dataSelect) Run() (err error) {
    var ferr = merr.Make("dataSelect.Run")
    
    dataFiles := sel.DataFiles
    if len(dataFiles) == 0 {
        return fmt.Errorf("At least one datafile must be specified!")
    }
    
    var startCheck, stopCheck checkerFun
    checker := func(s1zip *S1Zip) bool {
        return true
    }
    
    dateStart, dateStop := sel.DateStart, sel.DateStop

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
        
        ll, ur = sel.LowerLeft, sel.UpperRight
        
        aoi = AOI{
            Point{X: ll.Lon, Y: ll.Lat}, Point{X: ll.Lon, Y: ur.Lat},
            Point{X: ur.Lon, Y: ur.Lat}, Point{X: ur.Lon, Y: ll.Lat},
        }
    )
    
    writer := sel.OutFile
    defer writer.Close()
    
    for _, zip := range dataFiles {
        if s1zip, IWs, err = parseS1(zip.String(), sel.Pol, sel.CachePath);
           err != nil {
            return Handle(err,
                "failed to import S1Zip data from '%s'", zip)
        }
        
        if IWs.contains(aoi) && checker(s1zip) {
            writer.WriteString(s1zip.Path)
        }
    }
    
    if err = writer.Wrap(); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

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

func toZiplist(name string, one, two *S1Zip) (err error) {
    file := NewWriterFile(name)
    if err = file.Wrap(); err != nil {
        return
    }
    defer file.Close()
    
    if two == nil {
        file.WriteString(one.Path)
    } else {
        after := two.date.center.After(one.date.center)
        
        if after {
            file.WriteString(one.Path + "\n")
            file.WriteString(two.Path + "\n")
        } else {
            file.WriteString(two.Path + "\n")
            file.WriteString(one.Path + "\n")
        }
    }

    return file.Wrap()
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
