package sentinel1

import (
    "fmt"
    "log"
    "time"
    "strings"
    "path/filepath"

    "github.com/bozso/gotoolbox/cli/stream"
    
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/base"
    "github.com/bozso/gamma/date"
)

type SLC struct {
    nIW int
    Tab string
    time.Time
    IWs
}

func (s SLC) Date() time.Time {
    return s.Time
}

func FromTabfile(tab string) (s1 SLC, err error) {
    log.Printf("Parsing tabfile: '%s'.\n", tab)
    
    file, err := stream.Open(tab)
    if err != nil {
        return
    }
    defer file.Close()

    s1.nIW = 0
    
    scan := file.Scanner()
    
    for scan.Scan() {
        line := scan.Text()
        split := strings.Split(line, " ")
        
        log.Printf("Parsing IW%d\n", s1.nIW + 1)
        
        l := NewIW(split[0], split[1], split[2])
        
        
        s1.IWs[s1.nIW], err = l.Load()
        if err != nil {
            return
        }
        
        s1.nIW++
    }
    
    s1.Tab = tab
    
    s1.Time = s1.IWs[0].Date()
    return
}



func (s1 SLC) Move(dir string) (ms1 SLC, err error) {
    ms1 = s1
    
    newtab := filepath.Join(dir, filepath.Base(s1.Tab))
    
    file, err := stream.Create(newtab)
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
        
        line := fmt.Sprintf("%s %s %s\n",
            iw.DatFile, iw.ParFile, iw.TOPS_par)
        
        if _, err = file.WriteString(line); err != nil {
            return 
        }
    }
    
    return
}

func (s1 SLC) Exist() (b bool, err error) {
    for _, iw := range s1.IWs {
        if b, err = iw.Exist(); err != nil {
            err = utils.WrapFmt(err,
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

func (s1 SLC) Mosaic(out base.SLC, opts MosaicOpts) (err error) {
    opts.Looks.Default()
    
    bflg := 0
    
    if opts.BurstWindowFlag {
        bflg = 1
    }
    
    ref := "-"
    
    if len(opts.RefTab) == 0 {
        ref = opts.RefTab
    }
    
    _, err = mosaic.Call(s1.Tab, out.DatFile, out.ParFile, opts.Looks.Rng,
        opts.Looks.Azi, bflg, ref)
    
    return
}

var derampRef = common.Must("S1_deramp_TOPS_reference")

func (s1 SLC) DerampRef() (ds1 SLC, err error) {
    tab := s1.Tab
    
    if _, err = derampRef.Call(tab); err != nil {
        err = utils.WrapFmt(err, "failed to deramp reference SLC '%s'",
            tab)
        return
    }
    
    tab += ".deramp"
    
    if ds1, err = FromTabfile(tab); err != nil {
        err = utils.WrapFmt(err, "failed to import SLC from tab '%s'",
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
        err = utils.WrapFmt(err,
            "failed to deramp slave SLC '%s', reference: '%s'", tab, reftab)
        return
    }
    
    tab += ".deramp"
    
    if ret, err = FromTabfile(tab); err != nil {
        err = utils.WrapFmt(err, "failed to import SLC from tab '%s'",
            tab)
        return
    }
    
    return ret, nil
}

func (s1 SLC) RSLC(outDir string) (ret SLC, err error) {
    tab := strings.ReplaceAll(filepath.Base(s1.Tab), "SLC_tab", "RSLC_tab")
    tab = filepath.Join(outDir, tab)

    file, err := stream.Create(tab)
    if err != nil {
        return
    }
    defer file.Close()

    for ii := 0; ii < s1.nIW; ii++ {
        IW := s1.IWs[ii]
        
        dat := filepath.Join(outDir,
            strings.ReplaceAll(filepath.Base(IW.DatFile), "slc", "rslc"))
        par, TOPS_par := dat + ".par", dat + ".TOPS_par"
        
        ret.IWs[ii] = NewIW(dat, par, TOPS_par)
        
        //if err != nil {
            //err = DataCreateErr.Wrap(err, "IW")
            ////err = Handle(err, "failed to create new IW")
            //return
        //}
        
        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            return
        }
    }

    ret.Tab, ret.nIW = tab, s1.nIW

    return
}

var MLIFun = common.Select("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *SLC) MLI(mli *base.MLI, opt *base.MLIOpt) (err error) {
    opt.Parse()
    
    wflag := 0
    
    if opt.WindowFlag {
        wflag = 1
    }
    
    _, err = MLIFun.Call(s1.Tab, mli.DatFile, mli.ParFile,
        opt.Looks.Rng, opt.Looks.Azi, wflag, opt.RefTab)
    
    return
}

