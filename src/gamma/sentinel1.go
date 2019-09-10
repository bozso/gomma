package gamma

import (
	"fmt"
	"log"
    "time"
    // conv "strconv"
	fp "path/filepath"
	str "strings"
)


type (
    ByDate []S1Zip
    
    tplType int
    IWInfos [3]IWInfo
	
    S1Zip struct {
		path, zipBase, mission, dateStr, mode, productType, resolution string
		level, productClass, pol, absoluteOrbit, DTID, UID             string
		date                                                           date
	}
    
    IWInfo struct {
		num, nburst int
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
	IWAll = "[1-3]"
    
    tiff tplType = iota
    annot
    calib
    noise
    preview
    quicklook
)

var (
	burstCorner CmdFun
    s1exTemplate = "{{mission}}-iw{{iw}}-slc-{{pol}}-.*-{{abs_orb}}-" +
		"{{DTID}}-[0-9]{3}"
    
    s1templates = []string{
        tiff:  "measurement/%s.tiff",
        annot: "annotation/%s.xml",
        calib: "annotation/%s.xml",
        noise: "annotation/calibration/noise-%s.xml",
        preview: "preview/product-preview.html",
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

	self.mission = zipBase[:3]
	self.dateStr = zipBase[17:48]

	start, stop := zipBase[17:32], zipBase[33:48]

	if self.date, err = NewDate(long, start, stop); err != nil {
		return self,
			handle(err,
				"Could not create new date from strings: '%s' '%s'",
				start, stop)
	}

	self.mode = zipBase[4:6]
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

func (self *S1Zip) templates(types []tplType, pol, iw string) []string {
	// TODO: test
	
    root := self.path
    tpl := self.mainTemplate(pol, iw)
    
    ret := make([]string, len(types))
    
    for ii, tplType := range types {
        nextTpl := fmt.Sprintf(s1templates[tplType], tpl)
        
        ret[ii] = fp.Join(root, nextTpl)
    }
    
	return ret
}

func (self *S1Zip) Extract(names []tplType, ext extractInfo) ([]string, error) {
	templates := self.templates(names, ext.pol, ext.iw)

	ret, err := extract(self.path, ext.root, templates)

	if err != nil {
		return nil, fmt.Errorf(
			"In S1Zip.Extract: extraction of requested files failed!\nError: %w",
			err)
	}

	return ret, nil
}

func (self *S1Zip) useExtracted(ext extractInfo, types []tplType) ([]string, error) {
	templates := self.templates(types, ext.pol, ext.iw)
    
    ret, err := ext.filterFiles(templates)
    if err != nil {
        return nil, fmt.Errorf(
            "In S1Zip.useExtracted: Failed to filter extracted files!\nError: %w",
            err)
    }
    
    return ret, nil
}

// TODO: implement
func (self *S1Zip) IWInfo(ext extractInfo) (IWInfos, error) {
    handle := Handler("S1Zip.IWInfo")
    var (
        ret IWInfos
        iwNums = [3]string{"1", "2", "3"}
        types = []tplType{annot}
    )
    
    for ii, iw := range iwNums {
        templates := self.templates(types, ext.pol, iw)
        
        files, err := ext.filterFiles(templates)
        if err != nil {
            return ret, handle(err, "Filtering of extracted files failed!")
        }

        ret[ii], err = iwInfo(files[0])
        
        if err != nil {
            return ret, handle(err, "Failed to create IWInfo!")
        }
    }
    
    return ret, nil
}

func (self S1Zip) Date() time.Time {
    return self.date.center
}

func makePoint(info Params, max bool) (point, error) {
	handle := Handler("makePoint")

	var tpl_lon, tpl_lat string

	if max {
		tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
	} else {
		tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
	}

	x, err := info.Float(tpl_lon)

	if err != nil {
		return point{}, handle(err, "Could not get Longitude value!")
	}

	y, err := info.Float(tpl_lat)

	if err != nil {
		return point{}, handle(err, "Could not get Latitude value!")
	}

	return point{x: x, y: y}, nil
}

func iwInfo(path string) (IWInfo, error) {
	handle := Handler("iwInfo")
    var ret IWInfo
    
	//num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
	num := int(str.Split(path, "iw")[1][0])
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

	for ii := 0; ii < nburst; ii++ {
		tpl := fmt.Sprintf(burstTpl, ii)

		numbers[ii], err = TOPS.Float(tpl)

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

	return IWInfo{num: num, nburst: nburst, extent: rect{min: min, max: max},
		bursts: numbers}, nil
}

func (self *point) inIWs(IWs []IWInfo) bool {
	for _, iw := range IWs {
		if self.inRect(&iw.extent) {
			return true
		}
	}
	return false
}

func pointsInSLC(IWs []IWInfo, points [4]point) bool {
	sum := 0

	for _, point := range points {
		if point.inIWs(IWs) {
			sum++
		}
	}
	return sum == 4
}




func (self ByDate) Len() int { return len(self) }
func (self ByDate) Swap(i, j int) { self[i], self[j] = self[j], self[i] }

func (self ByDate) Less(i, j int) bool {
    return Before(&self[i], &self[j])
}

