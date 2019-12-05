package gamma

import (
    "os"
    "fmt"
    "log"
    "path/filepath"
    //str "strings"
)

type (
    S1Coreg struct {
        Tab, ID, OutDir, RslcPath, IfgPath string
        Hgt, Poly1, Poly2 string
        Looks RngAzi
        Clean, UseInter bool
        CoregOpt
    }
    
    CoregOut struct {
        RSLC SLC
        Rslc S1SLC
        Ifg IFG
        Ok bool
    }
)

var coregFun = Gamma.Must("S1_coreg_TOPS")

func (self *S1Coreg) Coreg(slc, ref *S1SLC) (ret CoregOut, err error) {
    ret.Ok = false
    cleaning, flag1 := 0, 0
    
    if self.Clean {
        cleaning = 1
    }
    
    if self.UseInter {
        flag1 = 1
    }
    
    slc1Tab, slc1ID := self.Tab, self.ID
    slc2Tab, slc2ID := slc.Tab, slc.Format(DateShort)
    
    // TODO: parse opt.hgt
    hgt := self.Hgt
    
    ret.Rslc, err = slc.RSLC(self.OutDir)
    
    if err != nil {
        err = Handle(err, "failed to create RSLC")
        return
    }
    
    exist, err := ret.Rslc.Exist()
    
    if err != nil {
        err = Handle(err, "failed to check whether target RSLC exists")
        return
    }
    
    if exist {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        
        if ret.Rslc, err = ret.Rslc.Move(self.RslcPath); err != nil {
            err = Handle(err, "failed to move '%s' to RSLC directory",
                ret.Rslc.Tab)
            return
        }
        
        ret.Ok = true
        
        return ret, nil
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'", slc2Tab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, ret.Rslc.Tab, hgt,
                          self.Looks.Rng, self.Looks.Azi, self.Poly1,
                          self.Poly2, self.CoherenceThresh,
                          self.FractionThresh, self.PhaseStdevThresh,
                          cleaning, flag1)
        
        if err != nil {
            err = Handle(err, "coregistration failed")
            return
        }
    } else {
        rslcRefTab, rslcRefID := ref.Tab, ref.Format(DateShort)
        
        log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
            rslcRefTab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, ret.Rslc.Tab, hgt,
                          self.Looks.Rng, self.Looks.Azi, self.Poly1,
                          self.Poly2, self.CoherenceThresh,
                          self.FractionThresh, self.PhaseStdevThresh,
                          cleaning, flag1, rslcRefTab, rslcRefID)
        
        if err != nil {
            err = Handle(err, "coregistration failed")
            return
        }
    }
    
    if ret.RSLC, err = NewSLC(slc2ID + ".rslc", ""); err != nil {
        err = Handle(err, "failed to create SLC struct")
        return
    }
    
    ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
    var ifg IFG
    if ifg, err = NewIFG(ID + ".diff", ID + ".off", ID + ".diff_par");
       err != nil {
        return
    }
    
    ifg.Quality = ID + ".results"
    
    if err != nil {
        err = Handle(err, "failed to create IFG '%s'", ID + ".diff")
        return
    }
    
    if ret.Ok, err = ifg.CheckQuality(); err != nil {
        err = Handle(err, "failed to check coregistration quality '%s'",
            ifg.Quality)
        return
    }
    
    if !ret.Ok {
        return ret, nil
    }
    
    if ret.Rslc, err = ret.Rslc.Move(self.RslcPath); err != nil {
        err = Handle(err, "failed to move '%s' to RSLC directory", ret.Rslc.Tab)
        return
    }
    
    if ret.Ifg, err = ifg.Move(self.IfgPath); err != nil {
        err = Handle(err, "failed to move interferogram '%s' to IFG directory",
            ifg.Dat)
        return
    }

    if self.Clean {
        var glob []string
        pattern := filepath.Join(self.OutDir, slc1ID + "*")
        
        if glob, err = filepath.Glob(pattern); err != nil {
            err = Handle(err, "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if err = os.Remove(file); err != nil {
                err = Handle(err, "failed to remove file '%s'", file)
                return
            }
        }
        
        pattern = filepath.Join(self.OutDir, slc2ID + "*")
        
        if glob, err = filepath.Glob(pattern); err != nil {
            err = Handle(err, "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if err = os.Remove(file); err != nil {
                err = Handle(err, "failed to remove file '%s'", file)
                return
            }
        }
    }
    
    return ret, nil
}
