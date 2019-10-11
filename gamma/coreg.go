package gamma

import (
    "os"
    "fmt"
    "log"
    fp "path/filepath"
    //str "strings"
)

type (
    S1Coreg struct {
        tab, ID, outDir, rslcPath, ifgPath string
        hgt, poly1, poly2 string
        Looks RngAzi
        clean, useInter bool
        coreg
    }
)

var coregFun = Gamma.must("S1_coreg_TOPS")

func (self *S1Coreg) Coreg(slc, ref *S1SLC) (ret bool, RSLC S1SLC, ifg IFG, err error) {
    ret = false
    cleaning, flag1 := 0, 0
    
    if self.clean {
        cleaning = 1
    }
    
    if self.useInter {
        flag1 = 1
    }
    
    slc1Tab, slc1ID := self.tab, self.ID
    slc2Tab, slc2ID := slc.tab, slc.Format(DateShort)
    
    // TODO: parse opt.hgt
    hgt := self.hgt
    
    RSLC, err = slc.RSLC(self.outDir)
    
    if err != nil {
        err = Handle(err, "failed to create RSLC")
        return
    }
    
    exist, err := RSLC.Exist()
    
    if err != nil {
        err = Handle(err, "failed to check whether target RSLC exists")
        return
    }
    
    if exist {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        
        err = RSLC.Move(self.rslcPath)
        
        if err != nil {
            err = Handle(err, "failed to move '%s' to RSLC directory", RSLC.tab)
            return
        }
        
        return true, RSLC, ifg, nil
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'", slc2Tab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, RSLC.tab, hgt,
                          self.Looks.Rng, self.Looks.Azi, self.poly1,
                          self.poly2, self.CoherenceThresh,
                          self.FractionThresh, self.PhaseStdevThresh,
                          cleaning, flag1)
        
        if err != nil {
            err = Handle(err, "coregistration failed")
            return
        }
    } else {
        rslcRefTab, rslcRefID := ref.tab, ref.Format(DateShort)
        
        log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
            rslcRefTab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, RSLC.tab, hgt,
                          self.Looks.Rng, self.Looks.Azi, self.poly1,
                          self.poly2, self.CoherenceThresh,
                          self.FractionThresh, self.PhaseStdevThresh,
                          cleaning, flag1, rslcRefTab, rslcRefID)
        
        if err != nil {
            err = Handle(err, "coregistration failed")
            return
        }
    }
    
    ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
    ifg, err = NewIFG(ID + ".diff", ID + ".off", "", ID + ".diff_par",
        ID + ".results")
    
    if err != nil {
        err = Handle(err, "failed to create IFG '%s'", ID + ".diff")
        return
    }
    
    ok, err := ifg.CheckQuality()
    
    if err != nil {
        err = Handle(err, "failed to check coregistration quality '%s'",
            ifg.quality)
        return
    }
    
    if !ok {
        return false, RSLC, ifg, nil
    }
    
    err = RSLC.Move(self.rslcPath)
    
    if err != nil {
        err = Handle(err, "failed to move '%s' to RSLC directory", RSLC.tab)
        return
    }
    
    err = ifg.Move(self.ifgPath)
    
    if err != nil {
        err = Handle(err, "failed to move interferogram '%s' to IFG directory",
            ifg.Dat)
        return
    }
    
    glob, err := fp.Glob(fp.Join(self.outDir, slc1ID + "*"))
    
    if err != nil {
        err = Handle(err, "globbing for leftover files from coregistration failed")
        return
    }
    
    for _, file := range glob {
        err = os.Remove(file)
        
        if err != nil {
            err = Handle(err, "failed to remove file '%s'", file)
            return
        }
    }
    
    
    glob, err = fp.Glob(fp.Join(self.outDir, slc2ID + "*"))
    
    if err != nil {
        err = Handle(err, "globbing for leftover files from coregistration failed")
        return
    }
    
    for _, file := range glob {
        err = os.Remove(file)
        
        if err != nil {
            err = Handle(err, "failed to remove file '%s'", file)
            return
        }
    }
    
    return true, RSLC, ifg, nil
}
