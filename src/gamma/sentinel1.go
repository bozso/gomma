package gamma

import (
	"fmt"
	"log"
	"time"
    "math"
	zip "archive/zip"
	fp "path/filepath"
	str "strings"
)

type (
	ByDate []S1Zip

	tplType int
	IWInfos [3]IWInfo

	S1Zip struct {
		path, zipBase, mission, dateStr, mode, productType, resolution string
		safe, level, productClass, pol, absoluteOrbit, DTID, UID       string
		date                                                           date
	}

	IWInfo struct {
		nburst int
		extent      rect
		bursts      [9]float64
	}

	S1IW struct {
		dataFile
		TOPS_par ParamFile
	}

	S1SLC struct {
		nIW int
		IWs [9]S1IW
		tab string
	}
)

const (
	burstTpl = "burst_asc_node_%d"
	IWAll    = "[1-3]"

	tiff tplType = iota
	annot
	calib
	noise
	preview
	quicklook
)

var (
	burstCorner  CmdFun
	s1exTemplate = "{{mission}}-iw{{iw}}-slc-{{pol}}-.*-{{abs_orb}}-" +
		"{{DTID}}-[0-9]{3}"

	s1templates = []string{
		tiff:      "measurement/%s.tiff",
		annot:     "annotation/%s.xml",
		calib:     "annotation/%s.xml",
		noise:     "annotation/calibration/noise-%s.xml",
		preview:   "preview/product-preview.html",
		quicklook: "preview/quick-look.png",
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

func NewS1Zip(zipPath string) (S1Zip, error) {

	var err error
	self := S1Zip{}
	handle := Handler("NewS1Zip")

	zipBase := fp.Base(zipPath)
	self.path, self.zipBase = zipPath, zipBase

	self.mission = str.ToLower(zipBase[:3])
	self.dateStr = zipBase[17:48]

	start, stop := zipBase[17:32], zipBase[33:48]

	if self.date, err = NewDate(long, start, stop); err != nil {
		return self,
			handle(err,
				"Could not create new date from strings: '%s' '%s'",
				start, stop)
	}

	self.mode = zipBase[4:6]
    self.safe = str.ReplaceAll(zipBase, ".zip", ".SAFE")
	self.productType = zipBase[7:10]
	self.resolution = string(zipBase[10])
	self.level = string(zipBase[12])
	self.productClass = string(zipBase[13])
	self.pol = zipBase[14:16]
	self.absoluteOrbit = zipBase[49:55]
	self.DTID = zipBase[56:62]
	self.UID = zipBase[63:67]

	return self, nil
}

func (self *S1Zip) mainTemplate(pol, iw string) string {
	const rexTemplate = "%s-iw%s-slc-%s-.*-%s-%s-[0-9]{3}"

	return fmt.Sprintf(rexTemplate, self.mission, iw, pol, self.absoluteOrbit,
		self.DTID)
}

func (self *S1Zip) template(mode tplType, pol, iw string) string {
	// TODO: test
	tpl := self.mainTemplate(pol, iw)

	return fp.Join(self.safe, fmt.Sprintf(s1templates[mode], tpl))
}


func (self *S1Zip) IWInfo(ext extractInfo) (IWInfos, error) {
	handle := Handler("S1Zip.IWInfo")
	var ret IWInfos
    
    pol, path := ext.pol, self.path
    zip, err := zip.OpenReader(path)
    
    if err != nil {
        return ret, handle(err, "Could not open zipfile: '%s'!", path)
    }
    
    for ii := 1; ii < 4; ii++ {
        template := self.template(annot, pol, fmt.Sprintf("%d", ii))
        extracted, err := ext.extract(zip, template)
        // log.Fatalf("%s", extracted)

        if err != nil {
            return ret, handle(err,
                "Failed to extract annotation file from '%s'!", path)
        }
        
        if ret[ii - 1], err = iwInfo(extracted); err != nil {
            return ret, handle(err,
                "Parsing of IW information of annotation file '%s' failed!",
                extracted)
            
        }
        
    }
	return ret, nil
}

func (self S1Zip) Date() time.Time {
	return self.date.center
}

func makePoint(info Params, max bool) (point, error) {
	handle := Handler("makePoint")

	var (
        tpl_lon, tpl_lat string
        ret point
        err error
    )

	if max {
		tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
	} else {
		tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
	}

	if ret.x, err = info.Float(tpl_lon); err != nil {
		return ret, handle(err, "Could not get Longitude value!")
	}

	if ret.y, err = info.Float(tpl_lat); err != nil {
		return ret, handle(err, "Could not get Latitude value!")
	}

	return ret, nil
}

func iwInfo(path string) (IWInfo, error) {
	handle := Handler("iwInfo")
	var ret IWInfo
    
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    
    if len(path) == 0 {
        return ret, handle(nil,
            "path to annotation file is an empty string: '%s'", path)
    }
    
	// Check(err, "Failed to retreive IW number from %s", path);

	par, err := TmpFile()

	if err != nil {
		return ret, handle(err, "Failed to create tmp file!")
	}

	TOPS_par, err := TmpFile()

	if err != nil {
		return ret, handle(err, "Failed to create tmp file!")
	}

	_info, err := Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par)
    log.Fatalf("info: %s", _info)
    
	if err != nil {
		return ret, handle(err, "Failed to parse parameter files!")
	}

	info := FromString(_info, ":")
	TOPS, err := FromFile(TOPS_par, ":")
    

	if err != nil {
		return ret, handle(err, "Could not parse TOPS_par file!")
	}

	nburst, err := TOPS.Int("number_of_bursts")

	if err != nil {
		return ret, handle(err, "Could not retreive number of bursts!")
	}

	var numbers [9]float64

	for ii := 1; ii < nburst + 1; ii++ {
		tpl := fmt.Sprintf(burstTpl, ii)

		numbers[ii - 1], err = TOPS.Float(tpl)

		if err != nil {
			return ret, handle(err, "Could not get burst number: '%s'",
				tpl)
		}
	}

	max, err := makePoint(info, true)

	if err != nil {
		return ret, handle(err, "Could not create max latlon point!")
	}

	min, err := makePoint(info, false)

	if err != nil {
		return ret, handle(err, "Could not create min latlon point!")
	}

	return IWInfo{nburst: nburst, extent: rect{min: min, max: max},
		bursts: numbers}, nil
}


func (self *point) inIWs(IWs IWInfos) bool {
	for _, iw := range IWs {
		if self.inRect(&iw.extent) {
			return true
		}
	}
	return false
}

func pointsInSLC(IWs IWInfos, points [4]point) bool {
	sum := 0

	for _, point := range points {
		if point.inIWs(IWs) {
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

func IWAbsDiff(one IWInfos, two IWInfos) (float64, error) {
    sum := 0.0
    
    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            return 0.0, fmt.Errorf(
            "In: IWInfos.AbsDiff: number of burst in first IW%d (%d) " +
            "is not equal to the number of burst in the second IW%d (%d)",
            nburst1, ii, nburst2, ii)
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
	return Before(&self[i], &self[j])
}
