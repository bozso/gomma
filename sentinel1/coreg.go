package sentinel1

import (
    "os"
    "fmt"
    "log"

    "github.com/bozso/gotoolbox/cli"
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/date"
    ifg "github.com/bozso/gomma/interferogram"
)


type (
    CoregOut struct {
        RSLC slc.SLC
        Rslc SLC
        Ifg ifg.File
    }
    
    CoregOpt struct {
        IfgPath, RslcPath               path.Dir
        Tab, ID                         string
        OutDir                          path.Dir
        Hgt, Poly1, Poly2, Mli          string
        CoherenceThresh, FractionThresh float64
        PhaseStdevThresh                float64
        Clean, UseInter                 bool  
        Looks                           common.RngAzi
    }
)

var coregFun = common.Must("S1_coreg_TOPS")

func (s1 *CoregOpt) SetCli(c *cli.Cli) {

    c.Var(&s1.IfgPath, "ifg", "Output interferogram metadata file")
        
    c.Var(&s1.RslcPath, "rslc", "Output RSLC metadata file")
    c.Var(&s1.OutDir, "outDir", "Output directory")

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

func (sc *CoregOpt) Coreg(Slc, ref *SLC) (c CoregOut, err error) {
    cleaning, flag1 := 0, 0
    
    if sc.Clean {
        cleaning = 1
    }
    
    if sc.UseInter {
        flag1 = 1
    }
    
    slc1Tab, slc1ID := sc.Tab, sc.ID
    slc2Tab, slc2ID := Slc.Tab, date.Short.Format(Slc)
    
    // TODO: parse opt.hgt
    hgt := sc.Hgt
    
    if c.Rslc, err = Slc.RSLC(sc.OutDir); err != nil {
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
    
    args := []interface{}{
        slc1Tab, slc1ID, slc2Tab, slc2ID, c.Rslc.Tab, hgt,
        sc.Looks.Rng, sc.Looks.Azi, sc.Poly1, sc.Poly2,
        sc.CoherenceThresh, sc.FractionThresh, sc.PhaseStdevThresh,
        cleaning, flag1,
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'", slc2Tab)
        
    } else {
        rslcRefTab, rslcRefID := ref.Tab, date.Short.Format(ref)
        
        log.Printf("Coregistering: '%s'. Reference: '%s'", slc2Tab,
            rslcRefTab)
        
        args = append(args, rslcRefTab, rslcRefID)
    }

    _, err = coregFun.Call(args...)
    if err != nil {
        return
    }
    
    if c.RSLC, err = slc.New(slc2ID + ".rslc", ""); err != nil {
        return
    }
    
    ID := path.New(fmt.Sprintf("%s_%s", slc1ID, slc2ID))
    
    loader := ifg.New(ID.AddExt("diff")).WithParFile(ID.AddExt("off"))
    loader = loader.WithDiffPar(ID.AddExt("diff_par"))
    loader = loader.WithQuality(ID.AddExt("results"))
    loader = loader.WithSimUnwrap(ID.AddExt("sim"))
    
    c.Ifg, err = loader.Load()
    if err != nil {
        return
    }
    
    if c.Rslc, err = c.Rslc.Move(sc.RslcPath); err != nil {
        return
    }
    
    if c.Ifg, err = c.Ifg.Move(sc.IfgPath); err != nil {
        err = errors.WrapFmt(err,
            "failed to move interferogram '%s' to IFG directory",
            c.Ifg.DatFile)
        return
    }

    if sc.Clean {
        glob, Err := sc.OutDir.Join(slc1ID + "*").Glob()
        
        if Err != nil {
            err = errors.WrapFmt(Err,
                "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if Err = os.Remove(file.String()); err != nil {
                err = errors.WrapFmt(Err, "failed to remove file '%s'",
                    file)
                return
            }
        }
        
        glob, Err = sc.OutDir.Join(slc2ID + "*").Glob()
        
        if Err != nil {
            err = errors.WrapFmt(Err,
                "globbing for leftover files from coregistration failed")
            return
        }
        
        for _, file := range glob {
            if err = os.Remove(file.String()); err != nil {
                err = errors.WrapFmt(err, "failed to remove file '%s'",
                    file)
                return
            }
        }
    }
    
    return
}
