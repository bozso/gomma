package sentinel1

import (
    "log"
    "time"
    "strings"
    "bufio"

    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/splitted"
    "github.com/bozso/gotoolbox/errors"
    
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/mli"
    "github.com/bozso/gomma/date"
)

type SLCMeta struct {
    nIW int
    time.Time
}

type SLC struct {
    Tab path.ValidFile
    SLCMeta
    IWs
}

func (s SLC) Date() time.Time {
    return s.Time
}

func FromTabfile(tab path.ValidFile) (s1 SLC, err error) {
    log.Printf("Parsing tabfile: '%s'.\n", tab)
    
    file, err := tab.Open()
    if err != nil {
        return
    }
    defer file.Close()

    s1.nIW = 0
    
    scan := bufio.NewScanner(file)
    
    for scan.Scan() {
        line := scan.Text()
        split, err := splitted.New(line, " ")
        if err != nil {
            return s1, err
        }
        
        log.Printf("Parsing IW%d\n", s1.nIW + 1)
        
        var f path.File
        
        err = split.ValueAt(&f, 0)
        if err != nil {
            return s1, err
        }
        
        l := NewIW(f)

        err = split.ValueAt(&f, 1)
        if err != nil {
            return s1, err
        }
        
        l = l.WithPar(f)

        err = split.ValueAt(&f, 2)
        if err != nil {
            return s1, err
        }
        
        l = l.WithTOPS(f)

        s1.IWs[s1.nIW], err = l.Load()
        if err != nil {
            return s1, err
        }
        
        s1.nIW++
    }
    
    if err = scan.Err(); err != nil {
        return
    }
    
    s1.Tab = tab
    
    s1.Time = s1.IWs[0].Date()
    return
}


func (s1 SLC) Move(dir path.Dir) (ms1 SLC, err error) {
    ms1 = s1
    
    newTab := dir.Join(s1.Tab.Base().GetPath())
    
    file, err := newTab.Create()
    if err != nil {
        return
    }
    defer file.Close()
    
    for ii := 0; ii < s1.nIW; ii++ {
        ms1.IWs[ii], err = s1.IWs[ii].Move(dir)
        if err != nil {
            return
        }
        
        iw := ms1.IWs[ii]
        
        if _, err = file.WriteString(iw.Tabline()); err != nil {
            return 
        }
    }
    
    ms1.Tab, err = newTab.ToFile().ToValid()
    return
}

func (s1 SLC) Exist() (b bool, err error) {
    for _, iw := range s1.IWs {
        if b, err = iw.Exist(); err != nil {
            err = errors.WrapFmt(err,
                "failed to determine whether IW datafile exists")
            return
        }

        if !b {
            return
        }
    }
    return true, nil
}

type MosaicOpts struct {
    Looks common.RngAzi
    BurstWindowFlag bool
    RefTab string
}

var mosaic = common.Must("SLC_mosaic_S1_TOPS")

func (s1 SLC) Mosaic(out slc.SLC, opts MosaicOpts) (err error) {
    opts.Looks.Default()
    
    bflg := 0
    
    if opts.BurstWindowFlag {
        bflg = 1
    }
    
    ref := "-"
    
    if len(opts.RefTab) == 0 {
        ref = opts.RefTab
    }
    
    _, err = mosaic.Call(s1.Tab, out.DatFile, out.ParFile,
        opts.Looks.Rng, opts.Looks.Azi, bflg, ref)
    
    return
}

var derampRef = common.Must("S1_deramp_TOPS_reference")

func (s1 SLC) DerampRef() (ds1 SLC, err error) {
    tab := s1.Tab
    
    if _, err = derampRef.Call(tab); err != nil {
        err = errors.WrapFmt(err, "failed to deramp reference SLC '%s'",
            tab)
        return
    }
    
    tab, err = tab.AddExt("deramp").ToFile().ToValid()
    if err != nil {
        return
    }
    
    if ds1, err = FromTabfile(tab); err != nil {
        err = errors.WrapFmt(err, "failed to import SLC from tab '%s'",
            tab)
        return
    }
    
    return ds1, nil
}

var derampSlave = common.Must("S1_deramp_TOPS_slave")

func (s1 SLC) DerampSlave(ref *SLC, looks common.RngAzi, keep bool) (ret SLC, err error) {
    looks.Default()
    
    reftab, tab, id := ref.Tab, s1.Tab, date.Short.Format(s1)
    
    clean := 1
    
    if keep {
        clean = 0
    }
    
    _, err = derampSlave.Call(tab, id, reftab, looks.Rng, looks.Azi,
        clean)
    
    if err != nil {
        err = errors.WrapFmt(err,
            "failed to deramp slave SLC '%s', reference: '%s'", tab, reftab)
        return
    }
    
    tab, err = tab.AddExt("deramp").ToFile().ToValid()
    if err != nil {
        return
    }
    
    if ret, err = FromTabfile(tab); err != nil {
        err = errors.WrapFmt(err, "failed to import SLC from tab '%s'",
            tab)
        return
    }
    
    return ret, nil
}

func (s1 SLC) RSLC(outDir path.Dir) (sp SLCPath, err error) {
    nIW := s1.nIW
    
    for ii := 0; ii < nIW; ii++ {
        dat := outDir.Join(strings.ReplaceAll(
            s1.IWs[ii].DatFile.Base().String(), "slc", "rslc"))
        
        sp.IWPaths[ii] = NewIW(dat.ToFile())
    }
    
    tab := strings.ReplaceAll(s1.Tab.Base().String(),
        "SLC_tab", "RSLC_tab")

    sp.SLCMeta, sp.Tab = s1.SLCMeta, outDir.Join(tab).ToFile()
    
    err = sp.CreateTabfile()
    if err != nil {
        return
    }
    
    return
}

var MLIFun = common.Select("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *SLC) MLI(mli *mli.MLI, opt *mli.Options) (err error) {
    opt.Parse()
    
    wflag := 0
    
    if opt.WindowFlag {
        wflag = 1
    }
    
    _, err = MLIFun.Call(s1.Tab, mli.DatFile, mli.ParFile,
        opt.Looks.Rng, opt.Looks.Azi, wflag, opt.RefTab)
    
    return
}

