package gamma

import (
    "os"
    "log"
    "fmt"
    "math"
    fp "path/filepath"
    str "strings"
)

const(
    nMaxBurst = 10
    nIW = 3
	burstTpl = "burst_asc_node_%d"
)

const (
	tiff tplType = iota
	annot
	calib
	noise
	preview
	quicklook
    
    nTemplate = 6
)

type(
	tplType int
	templates [nTemplate]string
	
	S1Zip struct {
		Path           string    `json:"path"`
        Root           string    `json:"root"`
        zipBase        string    `json:"-"`
        mission        string    `json:"-"`
        dateStr        string    `json:"-"`
        mode           string    `json:"-"`
        productType    string    `json:"-"`
        resolution     string    `json:"-"` 
		Safe           string    `json:"safe"`
        level          string    `json:"-"`
        productClass   string    `json:"-"`
        pol            string    `json:"-"`
        absoluteOrbit  string    `json:"-"`
        DTID           string    `json:"-"`
        UID            string    `json:"-"`
		Templates      templates `json:templates`
		date                     `json:"date"`
	}
    
    S1Zips []*S1Zip
	ByDate S1Zips
    
    IWInfo struct {
		nburst int
		extent Rect
		bursts [nMaxBurst]float64
	}
	
    IWInfos [nIW]IWInfo

	S1IW struct {
		dataFile
		TOPS_par Params
	}
    
    IWs [nIW]S1IW

	S1SLC struct {
		date
        IWs IWs
		tab string
	}
)

var (
	burstCorner  CmdFun

    burstCorners = Gamma.selectFun("ScanSAR_burst_corners",
                                   "SLC_burst_corners")
	calibPath = fp.Join("annotation", "calibration")
    
    fmtNeeded = [nTemplate]bool {
        tiff:       true,
        annot:      true,
        calib:      true,
        noise:      true,
        preview:    false,
        quicklook:  false,
    }
)

func init() {
	var ok bool
	
	if burstCorner, ok = Gamma["ScanSAR_burst_corners"]; !ok {
		if burstCorner, ok = Gamma["SLC_burst_corners"]; !ok {
			log.Fatalf("No Fun.")
		}
	}
}

func NewS1Zip(zipPath, root string) (self *S1Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"
    handle := Handler("NewS1Zip")

    zipBase := fp.Base(zipPath)
    self = &S1Zip{Path: zipPath, zipBase: zipBase}

    self.mission = str.ToLower(zipBase[:3])
    self.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    if self.date, err = NewDate(long, start, stop); err != nil {
        err = handle(err, "Could not create new date from strings: '%s' '%s'",
            start, stop)
        return
    }
    
    self.mode = zipBase[4:6]
    safe := str.ReplaceAll(zipBase, ".zip", ".SAFE")
    tpl := fmt.Sprintf(rexTemplate, self.mission)
    
    self.Templates = [6]string{
        tiff: fp.Join(safe, "measurement", tpl + ".tiff"),
        annot: fp.Join(safe, "annotation", tpl + ".xml"),
        calib: fp.Join(safe, calibPath, fmt.Sprintf("calibration-%s.xml", tpl)),
        noise: fp.Join(safe, calibPath, fmt.Sprintf("noise-%s.xml", tpl)),
        preview: fp.Join(safe, "preview", "product-preview.html"),
        quicklook: fp.Join(safe, "preview", "quick-look.png"),
    }

    self.Safe = safe
    self.Root = fp.Join(root, self.Safe)
    self.productType = zipBase[7:10]
    self.resolution = string(zipBase[10])
    self.level = string(zipBase[12])
    self.productClass = string(zipBase[13])
    self.pol = zipBase[14:16]
    self.absoluteOrbit = zipBase[49:55]
    self.DTID = str.ToLower(zipBase[56:62])
    self.UID = zipBase[63:67]

    return self, nil
}

func (self *S1Zip) Info(exto *ExtractOpt) (ret IWInfos, err error) {
    handle := Handler("S1Zip.Info")
    ext, err := self.newExtractor(exto)
    var _annot string
    
    if err != nil {
        err = handle(err, "Failed to create new S1Extractor!")
        return 
    }
    
    defer ext.Close()
    
    for ii := 1; ii < 4; ii++ {
        _annot, err = ext.extract(annot, ii)
        
        if err != nil {
            err = handle(err, "Failed to extract annotation file from '%s'!",
                self.Path)
            return
        }
        
        if ret[ii - 1], err = iwInfo(_annot); err != nil {
            err = handle(err,
                "Parsing of IW information of annotation file '%s' failed!",
                _annot)
            return
        }
        
    }
    return ret, nil
}

func NewIW(dat, par, TOPS_par string) (self S1IW, err error) {
    handle := Handler("NewS1SLC")
    
    self.dataFile, err = NewDataFile(dat, par)
    
    if err != nil {
        err = handle(err,
            "Failed to create DataFile with dat: '%s' and par '%s'!",
            dat, par)
        return
    }
    
    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }
    
    self.TOPS_par, err = NewGammaParam(TOPS_par)
    
    if err != nil {
        err = handle(err, "Failed to parse TOPS_parfile: '%s'!", TOPS_par)
        return
    }
    
    return self, nil
}

func (self *S1SLC) Exist() (ret bool, err error) {
    for _, iw := range self.IWs {
        exist, err = iw.Exist()
        
        if err != nil {
            err = fmt.Errorf("Could not determine whether IW datafile exists!")
            return
        }
        
        if !exist {
            return false, nil
        }
    }
    return true, nil
}

func makePoint(info Params, max bool) (ret Point, err error) {
	handle := Handler("makePoint")

	var tpl_lon, tpl_lat string

	if max {
		tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
	} else {
		tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
	}

	if ret.X, err = info.Float(tpl_lon); err != nil {
		err = handle(err, "Could not get Longitude value!")
        return
	}

	if ret.Y, err = info.Float(tpl_lat); err != nil {
		err = handle(err, "Could not get Latitude value!")
        return
	}

	return ret, nil
}

func iwInfo(path string) (ret IWInfo, err error) {
	handle := Handler("iwInfo")
    
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    
    if len(path) == 0 {
        err = handle(nil, "path to annotation file is an empty string: '%s'",
            path)
        return
    }
    
	// Check(err, "Failed to retreive IW number from %s", path);

	par, err := TmpFile()

	if err != nil {
		err = handle(err, "Failed to create tmp file!")
        return
	}

	TOPS_par, err := TmpFile()

	if err != nil {
		err = handle(err, "Failed to create tmp file!")
        return
	}

	_, err = Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par)
    
	if err != nil {
		err = handle(err, "Could not import parameter files from '%s'!",
            path)
        return
	}
    
    _info, err := burstCorners(par, TOPS_par)
    
	if err != nil {
		err = handle(err, "Failed to parse parameter files!")
        return
	}
    
	info := FromString(_info, ":")
	TOPS, err := FromFile(TOPS_par, ":")

	if err != nil {
		err = handle(err, "Could not parse TOPS_par file!")
        return
	}

	nburst, err := TOPS.Int("number_of_bursts")

	if err != nil {
		err = handle(err, "Could not retreive number of bursts!")
        return
	}

	var numbers [nMaxBurst]float64
    
	for ii := 1; ii < nburst + 1; ii++ {
		tpl := fmt.Sprintf(burstTpl, ii)
        
		numbers[ii - 1], err = TOPS.Float(tpl)

		if err != nil {
			err = handle(err, "Could not get burst number: '%s'", tpl)
            return
		}
	}

	max, err := makePoint(info, true)

	if err != nil {
		err = handle(err, "Could not create max latlon point!")
        return
	}

	min, err := makePoint(info, false)

	if err != nil {
		err = handle(err, "Could not create min latlon point!")
        return
	}

	return IWInfo{nburst: nburst, extent: Rect{Min: min, Max: max},
		bursts: numbers}, nil
}

func (self *Point) inIWs(IWs IWInfos) bool {
	for _, iw := range IWs {
		if self.InRect(&iw.extent) {
			//log.Printf("%v in %v", *self, iw.extent)
            return true
		}
        //log.Printf("%v not in %v", *self, iw.extent)
	}
	return false
}

func (self IWInfos) contains(aoi AOI) bool {
	sum := 0
    
	for _, point := range aoi {
		if point.inIWs(self) {
			sum++
		}
	}
	return sum == 4
}

func diffBurstNum(burst1, burst2 float64) int {
    dburst := burst1 - burst2
    diffSqrt := math.Sqrt(dburst)
    
    return int(dburst + 1.0 + (dburst / (0.001 + diffSqrt)) * 0.5)    
}

func checkBurstNum(one, two IWInfos) bool {
    for ii := 0; ii < 3; ii++ {
        if one[ii].nburst != two[ii].nburst {
            return true
        }
    }
    return false
}

func IWAbsDiff(one, two IWInfos) (float64, error) {
    sum := 0.0
    
    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            return 0.0, fmt.Errorf(
            "In: IWInfos.AbsDiff: number of burst in first SLC IW%d (%d) " +
            "is not equal to the number of burst in the second SLC IW%d (%d)",
            ii + 1, nburst1, ii + 1, nburst2)
        }
        
        for jj := 0; jj < nburst1; jj++ {
            dburst := one[ii].bursts[jj] - two[ii].bursts[jj]
            sum += dburst * dburst
        }
    }
    
    return math.Sqrt(sum), nil
}

func (self *S1Zip) SLC(pol string) (ret S1SLC, err error) {
    const mode = "slc"
    tab := self.tab(mode, pol)
    
    exist, err = Exist(tab)
    
    if err != nil {
        return
    }
    
    if !exist {
        err = fmt.Errorf("In SLC: tabfile '%s' does not exist!", tab)
        return
    }
    
    for ii := 0; ii < 4; ii++ {
        dat, par, TOPS_par := self.SLCNames(mode, pol, ii)
        ret.IWs[ii - 1], err = NewIW(dat, par, TOPS_par)
        
        if err != nil {
            err = handle(err, "Could not create new IW!")
            return
        }
    }
    
    ret.tab = tab
    
    return ret, nil
}

func (self *S1Zip) RSLC(pol string) (ret S1SLC, err error) {
    const mode = "rslc"
    tab := self.tab(mode, pol)
    
    exist, err = Exist(tab)
    
    if err != nil {
        return
    }
    
    if !exist {
        err = fmt.Errorf("In SLC: tabfile '%s' does not exist!", tab)
        return
    }
    
    for ii := 0; ii < 4; ii++ {
        dat, par, TOPS_par := self.SLCNames(mode, pol, ii)
        ret.IWs[ii - 1], err = NewIW(dat, par, TOPS_par)
        
        if err != nil {
            err = handle(err, "Could not create new IW!")
            return
        }
    }
    
    ret.tab = tab
    
    return ret, nil
}

func (self *S1Zip) SLCNames(mode, pol string, ii int) (dat, par, TOPS string) {
    slcPath := fp.Join(self.Root, mode)
    
    dat  := fp.Join(slcPath, fmt.Sprintf("iw%d_%s.slc", ii, pol))
    par  := dat + ".par"
    TOPS := dat + "TOPS_par"
    
    return
}


func (self *S1Zip) tabName(mode, pol, string) string {
    return fp.Join(self.Root, mode, fmt.Sprintf("%s.tab", pol))
}

func (self *S1Zip) ImportSLC(exto *ExtractOpt) (ret S1SLC, err error) {
    handle := Handler("S1Zip.SLC")
    var _annot, _calib, _tiff, _noise string
    ext, err := self.newExtractor(exto)
    
    if err != nil {
        err = handle(err, "Failed to create new S1Extractor!")
        return
    }
    
    defer ext.Close()
    
    path, pol := self.Path, exto.pol
    
    slcPath := fp.Join(self.Root, "slc")
    tab := self.tabName("slc", pol)
    file, err := os.Create(tab)
    
    if err != nil {
        err = handle(err, "Failed to open file: '%s'!", tab)
        return
    }
    
    defer file.Close()
    
    for ii := 1; ii < 4; ii++ {
        _annot, err = ext.extract(annot, ii)
        
        if err != nil {
            err = handle(err, "Failed to extract annotation file from '%s'!",
                path)
            return
        }
        
        _calib, err = ext.extract(calib, ii)
        
        if err != nil {
            err = handle(err, "Failed to extract calibration file from '%s'!",
                path)
            return
        }
        
        _tiff, err = ext.extract(tiff, ii)
        
        if err != nil {
            err = handle(err, "Failed to extract tiff file from '%s'!",
                path)
            return
        }
        
        _noise, err = ext.extract(noise, ii)
        
        if err != nil {
            err = handle(err, "Failed to extract noise file from '%s'!",
                path)
            return
        }
        
        err = os.MkdirAll(slcPath, os.ModePerm)
        
        if err != nil {
            err = handle(err, "Failed to create directory: %s!", slcPath)
            return
        }
        
        dat, par, TOPS_par := self.SLCNames("slc", pol, ii)
        
        _, err = Gamma["par_S1_SLC"](_tiff, _annot, _calib, _noise, par, dat,
            TOPS_par)
        if err != nil {
            err = handle(err, "Failed to import datafiles into gamma format!")
            return
        }
        
        ret.IWs[ii - 1], err = NewIW(dat, par, TOPS_par)
        if err != nil {
            err = handle(err, "Could not create new S1SLC!")
            return
        }
        
        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)
        
        _, err = file.WriteString(line)
        
        if err != nil {
            err = handle(err, "Failed to write line '%s' to file'%s'!",
                line, tab)
            return
        }
    }
    
    ret.tab = tab
    
    return ret, nil
}


func (self *S1Zip) RSLC(pol string) (tab string, exist bool) {
    tab = self.tab("rslc", pol)
    
    
    
    file, err := os.Create(tab)
    if err != nil {
        err = handle(err, "Failed to open file: '%s'!", tab)
        return
    }
    defer file.Close()
    
    for ii = 1; ii < 4; ii++ {
        dat, par TOPS_par := self.SLC("rslc", pol, ii)
        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)
        
        _, err = file.WriteString(line)
        
        if err != nil {
            err = handle(err, "Failed to write line '%s' to file'%s'!",
                line, tab)
            return
        }
    }
    return tab, true
}

func (self *S1Zip) Quicklook(exto *ExtractOpt) (ret string, err error) {
    handle := Handler("S1Zip.Quicklook")
    
    ext, err := self.newExtractor(exto)
    
    if err != nil {
        err = handle(err, "Failed to create new S1Extractor!")
        return
    }
    
    defer ext.Close()
    
    path := self.Path
    
    ret, err = ext.extract(quicklook, 0)
    
    if err != nil {
        err = handle(err, "Failed to extract annotation file from '%s'!",
            path)
        return
    }
    
    return ret, nil
}

func (self ByDate) Len() int      { return len(self) }
func (self ByDate) Swap(i, j int) { self[i], self[j] = self[j], self[i] }

func (self ByDate) Less(i, j int) bool {
    return Before(self[i], self[j])
}
