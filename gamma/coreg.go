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
    
    CoregOut struct {
        rslc S1SLC
        ifg IFG
        ok bool
    }
)

var coregFun = Gamma.must("S1_coreg_TOPS")

func (self *S1Coreg) Coreg(slc, ref *S1SLC) (ret CoregOut, err error) {
    ret.ok = false
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
    
    ret.rslc, err = slc.RSLC(self.outDir)
    
    if err != nil {
        err = Handle(err, "failed to create RSLC")
        return
    }
    
    exist, err := ret.rslc.Exist()
    
    if err != nil {
        err = Handle(err, "failed to check whether target RSLC exists")
        return
    }
    
    if exist {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        
        err = ret.rslc.Move(self.rslcPath)
        
        if err != nil {
            err = Handle(err, "failed to move '%s' to RSLC directory",
                ret.rslc.tab)
            return
        }
        
        ret.ok = true
        
        return ret, nil
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'", slc2Tab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, ret.rslc.tab, hgt,
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
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, ret.rslc.tab, hgt,
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
    
    ret.ifg, err = NewIFG(ID + ".diff", ID + ".off", "", ID + ".diff_par",
        ID + ".results")
    
    if err != nil {
        err = Handle(err, "failed to create IFG '%s'", ID + ".diff")
        return
    }
    
    ret.ok, err = ret.ifg.CheckQuality()
    
    if err != nil {
        err = Handle(err, "failed to check coregistration quality '%s'",
            ret.ifg.quality)
        return
    }
    
    if !ret.ok {
        return ret, nil
    }
    
    err = ret.rslc.Move(self.rslcPath)
    
    if err != nil {
        err = Handle(err, "failed to move '%s' to RSLC directory", ret.rslc.tab)
        return
    }
    
    err = ret.ifg.Move(self.ifgPath)
    
    if err != nil {
        err = Handle(err, "failed to move interferogram '%s' to IFG directory",
            ret.ifg.Dat)
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
    
    return ret, nil
}
