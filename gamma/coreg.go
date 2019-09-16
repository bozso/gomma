package gamma

import (
	//"os"
    "fmt"
	"log"
	//fp "path/filepath"
	//str "strings"
)

type (
    
	CoregOpt struct {
		coreg
		hgt, poly1, poly2 string
		RangeLooks, AzimuthLooks int
		clean, useInter bool
	}
)

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
			
			_, err = _coreg(slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.tab, hgt,
                            opt.RangeLooks, opt.AzimuthLooks, opt.poly1,
                            opt.poly2, opt.CoherenceThresh, opt.FractionThresh,
                            opt.PhaseStdevThresh, cleaning, flag1)
            
            if err != nil {
                err = handle(err, "Coregistration failed!")
                return
            }
		} else {
            rslcRefTab, rslcRefID := rslcRef.tab, date2str(rslcRef, short)
			log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
				rslcRefTab)
			
			_, err = _coreg(slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.tab, hgt,
                            opt.RangeLooks, opt.AzimuthLooks, opt.poly1,
                            opt.poly2, opt.CoherenceThresh, opt.FractionThresh,
                            opt.PhaseStdevThresh, cleaning, flag1,
                            rslcRefTab, rslcRefID)
            
            if err != nil {
                err = handle(err, "Coregistration failed!")
                return
            }
		}
	}
    
	ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
	ret, err = NewIFG(ID + ".diff", ID + ".off", "", ID + ".diff_par",
        ID + ".coreg_quality")
    
    if err != nil {
        err = handle(err, "Could not create IFG from '%s'!", ID + ".diff")
    }
    
    //with open("coreg.output", "wb") as f:
    //    f.write(out)
    ok, err := ret.CheckQuality()
    
    if err != nil {
        err = handle(err, "Could not check coregistration quality for '%s'!",
            ret.dat)
        return
    }
    
    if !ok {
        err = handle(err,"Coregistration of '%s' failed!", slc2Tab)
        return
    }
	
    //ifg.move(("dat", "par", "diff_par", "qual"), diff_dir)
    //ifg.raster(mli=master["MLI"])
	
	return ret, nil
}
