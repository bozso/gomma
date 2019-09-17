package gamma

import (
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
)

type(
	tplType int
	templates [6]string
	
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
