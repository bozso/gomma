package sentinel1

import (
    "os"
    "fmt"
    "log"

    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/mli"
    "github.com/bozso/gomma/geo"
    ifg "github.com/bozso/gomma/interferogram"
)

type CoregOut struct {
    RSLC slc.SLC
    Rslc SLC
    Ifg ifg.File
}

type BurstOverlapThresholds struct {
    // Coherence threshold in overlapping bursts
    Coherence              float64 `json:"coherence_thresh"`
    
    // Minimum fraction of coherent pixels in overlapping bursts
    MinCoherentPicFraction float64 `json:"min_coherent_pixels_fraction"`
    
    // Maximum allowed phase standard deviation
    MaxPhaseStdev         float64  `json:"max_phase_stdev"`
}

type Common struct {
    // Whether to clean temporary files
    Clean    bool           `json:"clean_temp_files"`
    
    // Wether to use previously calculated intermediate files
    UseInter bool           `json:"use_inter_files"`  
    
    Looks    common.RngAzi  `json:"looks"`    

    BurstOverlapThresh BurstOverlapThresholds `json:"burst_overlap_thresh"`
}

type CoregMeta struct {
    MasterMLI       path.ValidFile  `json:"master_mli"`
    MasterHeight    path.ValidFile  `json:"heights"`
    
    // Optional polynom file
    Poly1 path.File       `json:"poly1"`

    // Optional polynom file
    Poly2 path.File       `json:"poly2"`
    Common
}

func (cm CoregMeta) Parse() (co CoregOpt, err error) {
    if err = common.LoadJson(cm.MasterMLI, &co.MasterMLI); err != nil {
        return
    }

    if err = common.LoadJson(cm.MasterHeight, &co.MasterHeight); err != nil {
        return
    }
    
    co.Poly1, err = cm.Poly1.IfExists()
    if err != nil {
        return 
    }

    co.Poly2, err = cm.Poly2.IfExists()
    if err != nil {
        return 
    }
    
    co.Common = cm.Common
    return
}

type CoregOpt struct {
    MasterMLI       mli.MLI
    MasterHeight    geo.Height
    Poly1           *path.ValidFile
    Poly2           *path.ValidFile
    Common
}

var coregFun = common.Must("S1_coreg_TOPS")

func (sc *CoregOpt) Coreg(Slc, ref *SLC) (co CoregOut, err error) {
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
    
    rslc, err := Slc.RSLC(sc.OutDir)
    if err != nil {
        return
    }
    
    Rslc, err := rslc.Load()
    if err == nil {
        log.Printf("Coregistered RSLC already exists, moving it to directory.")
        co.Rslc, err = Rslc.Move(sc.RslcPath)
        return
    }
    
    bot := sc.BurstOverlapThresh
    
    args := []interface{}{
        slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.Tab, hgt,
        sc.Looks.Rng, sc.Looks.Azi, sc.Poly1, sc.Poly2,
        bot.Coherence, bot.MinCoherentPicFraction, bot.MaxPhaseStdev,
        cleaning, flag1,
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'.", slc2Tab)
    } else {
        rslcRefTab, rslcRefID := ref.Tab, date.Short.Format(ref)
        
        log.Printf(" Reference: '%s'.", rslcRefTab)
        
        args = append(args, rslcRefTab, rslcRefID)
    }

    _, err = coregFun.Call(args...)
    if err != nil {
        return
    }
    
    co.RSLC, err = slc.New(path.New(slc2ID).AddExt("rslc").ToFile()).Load()
    if err != nil {
        return
    }
    
    ID := path.New(fmt.Sprintf("%s_%s", slc1ID, slc2ID))
    
    loader := ifg.New(ID.AddExt("diff")).
        WithParFile(ID.AddExt("off")).
        WithDiffPar(ID.AddExt("diff_par")).
        WithQuality(ID.AddExt("results")).
        WithSimUnwrap(ID.AddExt("sim"))
    
    co.Ifg, err = loader.Load()
    if err != nil {
        return
    }
    
    if co.Rslc, err = co.Rslc.Move(sc.RslcPath); err != nil {
        return
    }
    
    if co.Ifg, err = co.Ifg.Move(sc.IfgPath); err != nil {
        err = errors.WrapFmt(err,
            "failed to move interferogram '%s' to IFG directory",
            co.Ifg.DatFile)
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
