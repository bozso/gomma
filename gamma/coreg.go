package gamma

import (
	//"os"
    "fmt"
	"log"
	fp "path/filepath"
	//str "strings"
)

type (
	S1Coreg struct {
        pol, tab, ID string
		hgt, poly1, poly2 string
		Looks RngAzi
		clean, useInter bool
		coreg
	}
)

var coregFun = Gamma.must("S1_coreg_TOPS")

func (self *S1Coreg) Coreg(slc, ref *S1Zip) (ret bool, err error) {
    pol := self.pol
    ret = false
    cleaning, flag1 := 0, 0
	
	if self.clean {
		cleaning = 1
	}
	
	if self.useInter {
		flag1 = 1
	}
    
    SLC, err := slc.SLC(pol)
    
    if err != nil {
        err = Handle(err, "Failed to retreive slave SLC!")
        return
    }
    
	slc1Tab, slc1ID := self.tab, self.ID
	slc2Tab, slc2ID := SLC.tab, date2str(slc, short)
	
    // TODO: parse opt.hgt
    hgt := self.hgt
    
    RSLC, err := slc.RSLC(pol)
    
    if err != nil {
        err = Handle(err, "Could not create RSLC!")
        return
    }
    
    exist, err := RSLC.Exist()
    
    if err != nil {
        err = Handle(err, "Could not check whether target RSLC exists!")
        return
    }
    
    if exist {
        return true, nil
    }
    
	if true {
		if ref == nil {
			log.Printf("Coregistering: '%s'", slc2Tab)
            
			_, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, RSLC.tab, hgt,
                              self.Looks.Rng, self.Looks.Azi, self.poly1,
                              self.poly2, self.CoherenceThresh,
                              self.FractionThresh, self.PhaseStdevThresh,
                              cleaning, flag1)
            
            if err != nil {
                err = Handle(err, "Coregistration failed!")
                return
            }
		} else {
            var REFSLC S1SLC
            REFSLC, err = ref.RSLC(pol)
            
            if err != nil {
                err = Handle(err, "Could not create RSLC!")
                return
            }
            
            rslcRefTab, rslcRefID := REFSLC.tab, date2str(ref, short)
			
            log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
				rslcRefTab)
			
			_, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, RSLC.tab, hgt,
                              self.Looks.Rng, self.Looks.Azi, self.poly1,
                              self.poly2, self.CoherenceThresh,
                              self.FractionThresh, self.PhaseStdevThresh,
                              cleaning, flag1, rslcRefTab, rslcRefID)
            
            if err != nil {
                err = Handle(err, "Coregistration failed!")
                return
            }
		}
	}
    
	ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
	ifg, err := NewIFG(ID + ".diff", ID + ".off", "", ID + ".diff_par",
        ID + ".coreg_quality")
    
    if err != nil {
        err = Handle(err, "Could not create IFG from '%s'!", ID + ".diff")
        return
    }
    
    ok, err := ifg.CheckQuality()
    
    if err != nil {
        err = Handle(err, "Could not check coregistration quality for '%s'!",
            ifg.quality)
        return
    }
	
    if !ok {
        return false, nil
    }
    
    err = ifg.Move(fp.Join(slc.Safe, "rslc"))
    
    if err != nil {
        err = Handle(err, "Could not move interferogram!")
        return
    }
    
	return true, nil
}
