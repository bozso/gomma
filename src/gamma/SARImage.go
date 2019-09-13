package gamma

import (
	"os"
    "fmt"
	"log"
	fp "path/filepath"
	str "strings"
)

type (

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
		date                      `json:"date"`
	}
    
    S1Zips []*S1Zip
	ByDate S1Zips
    
	CoregOpt struct {
		coreg
		hgt, poly1, poly2 string
		RangeLooks, AzimuthLooks int
		clean, useInter bool
	}
)

const (
	tiff tplType = iota
	annot
	calib
	noise
	preview
	quicklook
)

const (
	burstTpl = "burst_asc_node_%d"
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

func (self *S1Zip) SLC(pol string) (ret S1SLC, err error) {
	handle := Handler("S1Zip.SLC")
	
	slcPath := fp.Join(self.Root, "slc")
	tab := fp.Join(slcPath, fmt.Sprintf("%s.tab", pol))

	file, err := os.Create(tab)
	
	if err != nil {
		err = handle(err, "Failed to open file: '%s'!", tab)
		return
	}
	
	defer file.Close()
	
	for ii := 1; ii < 4; ii++ {
		dat      := fp.Join(slcPath, fmt.Sprintf("iw%d_%s.slc", ii, pol))
		par      := dat + ".par"
		TOPS_par := dat + "TOPS_par"
		
		line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)
		
		_, err = file.WriteString(line)
		
		if err != nil {
			err = handle(err, "Failed to write line '%s' to file'%s'!",
				line, tab)
			return
		}
		
		if ret.IWs[ii], err = NewIW(dat, par, TOPS_par); err != nil {
			err = handle(err, "Failed to parse IW%s of '%s'!", ii, self.Path)
			return
		}
	}
	
	ret.tab = tab
    ret.date = self.date
    
	return ret, nil
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
	tab := fp.Join(slcPath, fmt.Sprintf("%s.tab", exto.pol))
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
        
        dat      := fp.Join(slcPath, fmt.Sprintf("iw%d_%s.slc", ii, pol))
        par      := dat + ".par"
        TOPS_par := dat + "TOPS_par"
        
        _, err = Gamma["par_S1_SLC"](_tiff, _annot, _calib, _noise, par, dat,
            TOPS_par)
        
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
        
        if err != nil {
            err = handle(err, "Failed to import datafiles into gamma format!")
            return
        }
        
    }
	
	ret.tab = tab
	
	return ret, nil
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


var _coreg = Gamma["S1_coreg_TOPS"]

func S1Coreg(master, slc, rslc, rslcRef *S1SLC, opt CoregOpt) (ret IFG, err error) {
	handle := Handler("S1Coreg")
    cleaning, flag1 := 0, 0
	
	if opt.clean {
		cleaning = 1
	}
	
	if opt.useInter {
		flag1 = 1
	}

	slc1Tab, slc1ID := master.tab, date2str(master, short)
	slc2Tab, slc2ID := slc.tab, date2str(slc, short)
	
    // TODO: parse opt.hgt
    hgt := opt.hgt
    
	if true {
		if rslcRef == nil {
			log.Printf("Coregistering: '%s'", slc2Tab)
			
			_, err := _coreg(slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.tab, hgt,
				             opt.RangeLooks, opt.AzimuthLooks, opt.poly1,
							 opt.poly2, opt.CoherenceThresh, opt.FractionThresh,
				             opt.PhaseStdevThresh, cleaning, flag1)
            if err != nil {
                err = handle("Coregistration failed!")
                return
            }
		} else {
            rslcRefTab, rslcRefID := rslcRef.tab, date2str(rslcRef, short)
			log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
				rslcRefTab)
			
			_, err := _coreg(slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.tab, hgt,
							 opt.RangeLooks, opt.AzimuthLooks, opt.poly1,
							 opt.poly2, opt.CoherenceThresh, opt.FractionThresh,
							 opt.PhaseStdevThresh, cleaning, flag1,
							 rslcRefTab, rslcRefID)
            if err != nil {
                err = handle("Coregistration failed!")
                return
            }
		}
	}
    
	ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
	ret, err := NewIFG(ID + ".diff", ID + ".off", "", ID + ".diff_par",
        ID + ".coreg_quality")
    
    if err != nil {
        err = handle
    }
    
    //with open("coreg.output", "wb") as f:
    //    f.write(out)

    if ret.CheckQuality() {
        return handle(err,"Coregistration of '%s' failed!", slc2Tab)
    }
	
    //ifg.move(("dat", "par", "diff_par", "qual"), diff_dir)
    //ifg.raster(mli=master["MLI"])
	
	return nil
}
