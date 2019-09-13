package gamma

import (
	"os"
    "fmt"
	"log"
	"time"
    "math"
	zip "archive/zip"
	fp "path/filepath"
	str "strings"
)

type (
    S1Zips []*S1Zip
	ByDate S1Zips

	tplType int
	IWInfos [3]IWInfo

	S1Zip struct {
		Path           string `json:"path"`
        root           string `json:"root"`
        zipBase        string `json:"-"`
        mission        string `json:"-"`
        dateStr        string `json:"-"`
        mode           string `json:"-"`
        productType    string `json:"-"`
        resolution     string `json:"-"` 
		Safe           string `json:"safe"`
        level          string `json:"-"`
        productClass   string `json:"-"`
        pol            string `json:"-"`
        absoluteOrbit  string `json:"-"`
        DTID           string `json:"-"`
        UID            string `json:"-"`
		Dates          date   `json:"date"`
	}
    
    S1Extractor struct {
        ExtractOpt
        template string
        zip *zip.ReadCloser
    }
    
	IWInfo struct {
		nburst int
		extent      Rect
		bursts      [nMaxBurst]float64
	}

	S1IW struct {
		dataFile
		TOPS_par Params
	}

	S1SLC struct {
		// nIW int
		IWs [3]S1IW
		tab string
	}
)

const (
	burstTpl = "burst_asc_node_%d"
	IWAll    = "[1-3]"
    
    nMaxBurst = 10
    
	tiff tplType = iota
	annot
	calib
	noise
	preview
	quicklook
)

var (
	burstCorner  CmdFun

	s1templates = []string{
		tiff:      "measurement/%s.tiff",
		annot:     "annotation/%s.xml",
		calib:     "annotation/%s.xml",
		noise:     "annotation/calibration/noise-%s.xml",
		preview:   "preview/product-preview.html",
		quicklook: "preview/quick-look.png",
	}
    
    burstCorners = Gamma.selectFun("ScanSAR_burst_corners",
                                   "SLC_burst_corners")
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
	handle := Handler("NewS1Zip")

	zipBase := fp.Base(zipPath)
	self = &S1Zip{Path: zipPath, zipBase: zipBase}

	self.mission = str.ToLower(zipBase[:3])
	self.dateStr = zipBase[17:48]

	start, stop := zipBase[17:32], zipBase[33:48]

	if self.Dates, err = NewDate(long, start, stop); err != nil {
		err = handle(err, "Could not create new date from strings: '%s' '%s'",
            start, stop)
        return
	}

	self.mode = zipBase[4:6]
    self.Safe = str.ReplaceAll(zipBase, ".zip", ".SAFE")
    self.root = fp.Join(root, self.Safe)
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

func (self *S1Zip) template(mode tplType, pol, iw string) string {
	const rexTemplate = "%s-iw%s-slc-%s-.*-%s-%s-[0-9]{3}"

	tpl := fmt.Sprintf(rexTemplate, self.mission, iw, pol, self.absoluteOrbit,
		self.DTID)

	return fp.Join(self.Safe, fmt.Sprintf(s1templates[mode], tpl))
}


func (self *S1Zip) newExtractor(ext *ExtractOpt) (S1Extractor, error) {
	handle := Handler("S1Zip.extractor")
    const rexTemplate = "%s-iw%%d-slc-%%s-.*-%s-%s-[0-9]{3}"
    var err error
    ret := S1Extractor{}
    
    path := self.Path
    
    tpl := fmt.Sprintf(rexTemplate, self.mission, self.absoluteOrbit,
        self.DTID)
    
    ret.template = fp.Join(self.Safe, tpl)
    ret.pol      = ext.pol
    ret.root     = ext.root
    ret.zip, err = zip.OpenReader(path)

    if err != nil {
        return ret, handle(err, "Could not open zipfile: '%s'!", path)
    }
    
    return ret, nil
}

func (self *S1Extractor) extract(mode tplType, iw int) (string, error) {
    handle := Handler("S1Extractor.extract")
    tpl := fmt.Sprintf(self.template, iw, self.pol)
    tpl = fmt.Sprintf(s1templates[mode], tpl)
    
    ret, err := extract(self.zip, tpl, self.root)
    
    if err != nil {
        return "", handle(err, "Error occurred while extracting!")
    }
    
    return ret, nil
}

func (self *S1Extractor) Close() {
    self.zip.Close()
}

func (self *S1Zip) Info(exto *ExtractOpt) (ret IWInfos, err error) {
	handle := Handler("S1Zip.IWInfo")
    ext, err := self.newExtractor(exto)
    
    if err != nil {
        err = handle(err, "Failed to create new S1Extractor!")
        return 
    }
    
    defer ext.Close()
    
    for ii := 1; ii < 4; ii++ {
        var _annot string
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

func (self *S1Zip) SLC(exto *ExtractOpt) (ret S1SLC, err error) {
	handle := Handler("S1Zip.SLC")
    var ext S1Extractor
    ext, err = self.newExtractor(exto)
    
    if err != nil {
        err = handle(err, "Failed to create new S1Extractor!")
        return
    }
    
    defer ext.Close()
    
    path, pol := self.Path, exto.pol
    
    for ii := 1; ii < 4; ii++ {
        var _annot, _calib, _tiff, _noise string
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
        
        
        slcPath := fp.Join(self.root, "slc")
        
        err = os.MkdirAll(slcPath, os.ModePerm)
        
        if err != nil {
            err = handle(err, "Failed to create directory: %s!", slcPath)
            return
        }
        
        dat      := fp.Join(slcPath, fmt.Sprintf("iw%d_%s.slc", ii, pol))
        par      := dat + ".par"
        TOPS_par := dat + "TOPS_par"
        
        _, err = Gamma["par_S1_SLC"](_tiff, _annot, _calib, _noise, par, dat,
            TOPS_par)
        
        ret.IWs[ii - 1], err = NewS1SLC(dat, par, TOPS_par)
        
        if err != nil {
            err = handle(err, "Could not create new S1SLC!")
            return
        }
        
        if err != nil {
            err = handle(err, "Failed to import datafiles into gamma format!")
            return
        }
        
    }
	return ret, nil
}

func NewS1SLC(dat, par, TOPS_par string) (self S1IW, err error) {
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

func (self *S1Zip) Quicklook(exto *ExtractOpt) (ret string, err error) {
    handle := Handler("S1Zip.Quicklook")
    
    var ext S1Extractor
    ext, err = self.newExtractor(exto)
    
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

func (self S1Zip) Date() time.Time {
	return self.Dates.center
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


func (self ByDate) Len() int      { return len(self) }
func (self ByDate) Swap(i, j int) { self[i], self[j] = self[j], self[i] }

func (self ByDate) Less(i, j int) bool {
	return Before(self[i], self[j])
}
