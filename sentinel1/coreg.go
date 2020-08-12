package sentinel1

import (
    "os"
    "fmt"
    "log"

    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/math"

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
    Coherence              ifg.Coherence `json:"coherence"`
    
    // Minimum fraction of coherent pixels in overlapping bursts
    MinCoherentPixFraction math.Fraction `json:"min_coherent_pixels_fraction"`
    
    // Maximum allowed phase standard deviation
    MaxPhaseStdev          float64 `json:"max_phase_stdev"`
}

type OutPaths struct {
    RSLC path.EmptyDir `json:"rslc"`
    Ifg  path.EmptyDir `json:"ifg"`
}

type Common struct {
    // Whether to clean temporary files
    Clean    bool           `json:"clean_temp_files"`
    
    // Whether to use previously calculated intermediate files
    UseInter bool           `json:"use_inter_files"`  
    
    Looks    common.RngAzi  `json:"looks"`    

    BurstOverlapThresh BurstOverlapThresholds `json:"burst_overlap_thresh"`
    
    OutPaths OutPaths `json:"out_paths"`
}

type MasterFilePaths struct {
    MLI    path.ValidFile `json:"master_mli"`
    Height path.ValidFile `json:"height"`
    SLC    path.ValidFile `json:"slc"`
}

type MasterFiles struct {
    MLI    mli.MLI    `json:"mli"`
    Height geo.Height `json:"height"`
    SLC    SLC        `json:"slc"`
}

func (m MasterFilePaths) Parse() (mf MasterFiles, err error) {
    if err = common.LoadJson(m.MLI, &mf.MLI); err != nil {
        return
    }

    if err = common.LoadJson(m.Height, &mf.Height); err != nil {
        return
    }
    
    err = common.LoadJson(m.SLC, &mf.SLC)
    return
}

type CoregMeta struct {
    Master MasterFilePaths `json:"master_filepaths"`
    
    // Optional polynom file
    Poly1  path.File       `json:"poly1"`

    // Optional polynom file
    Poly2  path.File       `json:"poly2"`
    Common
}

func (cm CoregMeta) Parse() (co CoregOpt, err error) {
    if co.Master, err = cm.Master.Parse(); err != nil {
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
    Master MasterFiles
    Poly1  *path.ValidFile
    Poly2  *path.ValidFile
    Common
}

var coregFun = common.Must("S1_coreg_TOPS")

func (c *CoregOpt) Coreg(Slc, ref *SLC) (co CoregOut, err error) {
    cleaning, flag1 := 0, 0
    
    if c.Clean {
        cleaning = 1
    }
    
    if c.UseInter {
        flag1 = 1
    }
    
    m := c.Master.SLC
    
    slc1Tab, slc1ID := m.Tab, date.Short.Format(m)
    slc2Tab, slc2ID := Slc.Tab, date.Short.Format(Slc)
    
    // TODO: parse opt.hgt
    hgt := c.Master.Height
    
    rslc, err := Slc.RSLC(c.OutPaths.RSLC.Dir)
    if err != nil {
        return
    }
    
    bot := c.BurstOverlapThresh
    
    args := []interface{}{
        slc1Tab, slc1ID, slc2Tab, slc2ID, rslc.Tab, hgt,
        c.Looks.Rng, c.Looks.Azi, c.Poly1, c.Poly2,
        bot.Coherence, bot.MinCoherentPixFraction, bot.MaxPhaseStdev,
        cleaning, flag1,
    }
    
    if ref == nil {
        log.Printf("Coregistering: '%s'.", slc2Tab)
    } else {
        rslcRefTab, rslcRefID := ref.Tab, date.Short.Format(ref)
        
        log.Printf(" Reference: '%s'.\n", rslcRefTab)
        
        args = append(args, rslcRefTab, rslcRefID)
    }

    _, err = coregFun.Call(args...)
    if err != nil {
        return
    }
    
    co.RSLC, err = slc.New(
        path.New(slc2ID).AddExt("rslc").ToFile()).Load()
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
    
    if co.Rslc, err = co.Rslc.Move(c.OutPaths.RSLC.Dir); err != nil {
        return
    }
    
    if co.Ifg, err = co.Ifg.Move(c.OutPaths.Ifg.Dir); err != nil {
        err = errors.WrapFmt(err,
            "failed to move interferogram '%s' to IFG directory",
            co.Ifg.DatFile)
        return
    }

    if c.Clean {
        glob, Err := path.New(slc1ID + "*").Glob()
        
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
        
        glob, Err = path.New(slc2ID + "*").Glob()
        
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
