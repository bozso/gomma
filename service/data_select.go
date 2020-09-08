package service

import (
    "fmt"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/emath/geometry"
    
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/common"
    s1 "github.com/bozso/gomma/sentinel1"
)

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

func (s *S1Implement) SelectFiles(ss *SentinelSelect) (err error) {
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
        s1zip, IWs, err := parseS1(zip, s.CacheDir)
        
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

//func (s *selector) extOpt(satellite string) *ExtractOpt {
    //return &ExtractOpt{pol: s.Pol, 
        //root: filepath.Join(s.CachePath, satellite)}
//}


/*
type dataImporter struct {
    GeneralOpt
    IWs        [3]IMinmax `cli:"iws" usage:"IW burst indices"`
    
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
    date1 := date.Short.Format(s1.Date())
    
    for _, zip := range zips {
        if date1 == date.Short.Format(zip.Date())  && s1.Path != zip.Path {
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
