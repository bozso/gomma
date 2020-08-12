package server

type GeneralOpt struct {
    OutputDir, Pol, MasterDate, CachePath  string
    Looks      common.RngAzi
    InFile     stream.In
    OutFile    stream.Out
}

func (g *GeneralOpt) SetCli(c *cli.Cli) {
    g.InFile.SetCli(c, "infile", "Input file.")
    g.OutFile.SetCli(c, "outfile", "Input file.")
    
    c.StringVar(&g.OutputDir, "out", ".", "Output directory")
    c.StringVar(&g.Pol, "pol", "vv", "Polarisation used.")
    c.StringVar(&g.MasterDate, "masterDate", "", "")
    c.StringVar(&g.CachePath, "cachePath", "", "Cache path.")
    c.Var(&g.Looks, "looks", "Range, azimuth looks.")
}

type DataSelect struct {
    Files  []path.ValidFile
    DateStart  string
    DateStop   string
    AOI        common.AOI
    CheckZips  bool
    
}

type dataSelect struct {
    GeneralOpt
    DataFiles  cli.Paths `cli:"d,data" usage:"List of datafiles to process"`
    DateStart  string   `cli:"b,start" usage:"Start of date range"`
    DateStop   string   `cli:"e,stop" usage:"End of date range"`
    LowerLeft  common.LatLon   `cli:"ll,lowerLeft" usage:"Rectangle coordinates"`
    UpperRight common.LatLon   `cli:"ur,upperRight" usage:"Rectangle coordinates"`
    CheckZips  bool     `cli:"c,checkZips" usage:"Check zipfile integrity"`  
}

func (d *dataSelect) SetCli(c *cli.Cli) {
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
    dataFiles := sel.DataFiles
    if len(dataFiles) == 0 {
        return fmt.Errorf("At least one datafile must be specified!")
    }
    
    var startCheck, stopCheck checkerFun
    checker := func(s1zip *s1.Zip) bool {
        return true
    }
    
    dateStart, dateStop := sel.DateStart, sel.DateStop

    if len(dateStart) != 0 {
        _dateStart, err := date.Short.Parse(dateStart)
        
        if err != nil {
            return err
        }
        
        startCheck = func(s1zip *s1.Zip) bool {
            return s1zip.Start().After(_dateStart)
        }
    }
    
    if len(dateStop) != 0 {
        _dateStop, err := date.Short.Parse(dateStop)
        
        if err != nil {
            return err
        }
        
        stopCheck = func(s1zip *s1.Zip) bool {
            return s1zip.Stop().Before(_dateStop)
        }
    }
    
    if startCheck != nil && stopCheck != nil {
        checker = func(s1zip *s1.Zip) bool {
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
    
    
    ll, ur := sel.LowerLeft, sel.UpperRight
    
    aoi := common.AOI{
        common.Point{X: ll.Lon, Y: ll.Lat}, common.Point{X: ll.Lon, Y: ur.Lat},
        common.Point{X: ur.Lon, Y: ur.Lat}, common.Point{X: ur.Lon, Y: ll.Lat},
    }
    
    writer := sel.OutFile
    defer writer.Close()
    
    for _, zip := range dataFiles {
        s1zip, IWs, err := parseS1(zip, sel.Pol, sel.CachePath)
        
        if err != nil {
            return err
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
