package gamma

import (
    "os"
    "fmt"
    "log"
    "path/filepath"
    //str "strings"
)

type (
    S1CoregOut struct {
        RSLC SLC
        Rslc S1SLC
        Ifg IFG
    }
    
    S1CoregOpt struct {
        Tab, ID          string
        IfgPath          string  `cli:"i,ifg" usage:"Output interferogram metafile"`
        RslcPath         string  `cli:"r,rslc" usage:"Output RSLC metafile"`
        OutDir           string  `cli:"o,outdir" usage:"Output directory"`
        Hgt              string  `cli:"h,hgt" usage:""`
        Poly1            string  `cli:"p1,poly1" usage:""`
        Poly2            string  `cli:"p2,poly2" usage:""`
        Looks            RngAzi  `cli:"l,looks" usage:""`
        Clean            bool    `cli:"c,clean" usage:""`
        UseInter         bool    `cli:"u,useInter" usage:""`
        CoherenceThresh  float64 `cli:"c,coh"    dft:"0.8"`
        FractionThresh   float64 `cli:"f,frac"   dft:"0.01"`
        PhaseStdevThresh float64 `cli:"p,phase"  dft:"0.8"`
        Mli              string  `cli:"mli"`
    }
)

var coregFun = Gamma.Must("S1_coreg_TOPS")

func (sc *S1CoregOpt) Coreg(slc, ref *S1SLC) (c S1CoregOut, err error) {
    cleaning, flag1 := 0, 0
    
    if sc.Clean {
        cleaning = 1
    }
    
    if sc.UseInter {
        flag1 = 1
    }
    
    slc1Tab, slc1ID := sc.Tab, sc.ID
    slc2Tab, slc2ID := slc.Tab, slc.Format(DateShort)
    
    // TODO: parse opt.hgt
    hgt := sc.Hgt
    
    if c.Rslc, err = slc.RSLC(sc.OutDir); err != nil {
        return
    }
    
    exist := false
    if exist, err = c.Rslc.Exist(); err != nil {
        return
    }
    
    if exist {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        
        if c.Rslc, err = c.Rslc.Move(sc.RslcPath); err != nil {
            return
        }
        
        return c, nil
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'", slc2Tab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, c.Rslc.Tab,
                          hgt, sc.Looks.Rng, sc.Looks.Azi, sc.Poly1,
                          sc.Poly2, sc.CoherenceThresh,
                          sc.FractionThresh, sc.PhaseStdevThresh,
                          cleaning, flag1)
        
        if err != nil {
            return
        }
    } else {
        rslcRefTab, rslcRefID := ref.Tab, ref.Format(DateShort)
        
        log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
            rslcRefTab)
        
        _, err = coregFun(slc1Tab, slc1ID, slc2Tab, slc2ID, c.Rslc.Tab,
                          hgt, sc.Looks.Rng, sc.Looks.Azi, sc.Poly1,
                          sc.Poly2, sc.CoherenceThresh,
                          sc.FractionThresh, sc.PhaseStdevThresh,
                          cleaning, flag1, rslcRefTab, rslcRefID)
        
        if err != nil {
            return
        }
    }
    
    if c.RSLC, err = NewSLC(slc2ID + ".rslc", ""); err != nil {
        return
    }
    
    ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
    var ifg IFG
    if ifg, err = NewIFG(ID + ".diff", ID + ".off", ID + ".diff_par");
       err != nil {
        return
    }
    
    ifg.Quality = ID + ".results"
        
    if c.Rslc, err = c.Rslc.Move(sc.RslcPath); err != nil {
        return
    }
    
    if c.Ifg, err = ifg.Move(sc.IfgPath); err != nil {
        err = Handle(err, "failed to move interferogram '%s' to IFG directory",
            ifg.Dat)
        return
    }

    if sc.Clean {
        var glob []string
        pattern := filepath.Join(sc.OutDir, slc1ID + "*")
        
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
        
        pattern = filepath.Join(sc.OutDir, slc2ID + "*")
        
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
    
    return c, nil
}
