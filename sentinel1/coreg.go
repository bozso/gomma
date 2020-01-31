package sentinel1

import (
    "os"
    "fmt"
    "log"
    "path/filepath"
)

type (
    S1CoregOut struct {
        RSLC SLC
        Rslc sentinel1.S1SLC
        Ifg IFG
    }
    
    S1CoregOpt struct {
        Tab, ID, IfgPath, RslcPath      string
        OutDir, Hgt, Poly1, Poly2, Mli  string
        CoherenceThresh, FractionThresh float64
        PhaseStdevThresh                float64
        Clean, UseInter                 bool  
        Looks                           RngAzi
    }
)

var coregFun = Gamma.Must("S1_coreg_TOPS")

func (s1 *S1CoregOpt) SetCli(c *Cli) {

    c.StringVar(&s1.IfgPath, "ifg",
        "Output interferogram metadata file", "")
        
    c.StringVar(&s1.RslcPath, "rslc", "", "Output RSLC metadata file")
    c.StringVar(&s1.OutDir, "outDir", "", "Output directory")

    c.StringVar(&s1.Poly1, "poly1", "", "Polynom 1")
    c.StringVar(&s1.Poly2, "poly2", "", "Polynom 2")
    
    c.Var(&s1.Looks, "looks", "Number of looks.")
    
    c.BoolVar(&s1.Clean, "clean", false, "Cleanup temporary files.")
    c.BoolVar(&s1.UseInter, "useInter", false, "Use intermediate files.")
    
    c.Float64Var(&s1.CoherenceThresh, "cohThresh", 0.8,
        "Coherence threshold in overlapping bursts.")

    c.Float64Var(&s1.FractionThresh, "fracThresh", 0.01,
        "Fraction of coherent pixels in overlapping bursts.")

    c.Float64Var(&s1.PhaseStdevThresh, "stdThresh", 0.8,
        "Maximum allowed phase standard deviation.")
    
    c.StringVar(&s1.Mli, "mli", "", "Output? MLI metadata file.")
}

func (sc *S1CoregOpt) Coreg(slc, ref *S1SLC) (c S1CoregOut, err error) {
    var ferr = merr.Make("S1CoregOpt.Coreg")
    
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
        err = ferr.Wrap(err)
        return
    }
    
    exist := false
    if exist, err = c.Rslc.Exist(); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if exist {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        
        if c.Rslc, err = c.Rslc.Move(sc.RslcPath); err != nil {
            err = ferr.Wrap(err)
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
            err = ferr.Wrap(err)
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
            err = ferr.Wrap(err)
            return
        }
    }
    
    if c.RSLC, err = NewSLC(slc2ID + ".rslc", ""); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    ID := fmt.Sprintf("%s_%s", slc1ID, slc2ID)
    
    var ifg IFG
    if ifg, err = NewIFG(ID + ".diff", ID + ".off", ID + ".diff_par");
       err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    ifg.Quality = ID + ".results"
        
    if c.Rslc, err = c.Rslc.Move(sc.RslcPath); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if c.Ifg, err = ifg.Move(sc.IfgPath); err != nil {
        err = ferr.WrapFmt(err,
            "failed to move interferogram '%s' to IFG directory",
            ifg.Dat)
        return
    }

    if sc.Clean {
        var glob []string
        pattern := filepath.Join(sc.OutDir, slc1ID + "*")
        
        if glob, err = filepath.Glob(pattern); err != nil {
            err = ferr.WrapFmt(err,
                "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if err = os.Remove(file); err != nil {
                err = ferr.WrapFmt(err, "failed to remove file '%s'",
                    file)
                return
            }
        }
        
        pattern = filepath.Join(sc.OutDir, slc2ID + "*")
        
        if glob, err = filepath.Glob(pattern); err != nil {
            err = ferr.WrapFmt(err,
                "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if err = os.Remove(file); err != nil {
                err = ferr.WrapFmt(err, "failed to remove file '%s'",
                    file)
                return
            }
        }
    }
    
    return c, nil
}
