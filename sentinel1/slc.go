package sentinel1

import (
    "fmt"
    "os"
    "log"
    "time"
    "strings"
    "path/filepath"

    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/base"
)

type S1SLC struct {
    nIW int
    Tab string
    time.Time
    IWs
}

func FromTabfile(tab string) (s1 S1SLC, err error) {
    ferr := merr.Make("FromTabfile")
    
    log.Printf("Parsing tabfile: '%s'.\n", tab)
    
    file, err := utils.NewReader(tab)
    if err != nil {
        return
    }
    defer file.Close()

    s1.nIW = 0
    
    for file.Scan() {
        line := file.Text()
        split := strings.Split(line, " ")
        
        log.Printf("Parsing IW%d\n", s1.nIW + 1)
        
        s1.IWs[s1.nIW] = NewIW(split[0], split[1], split[2])
        s1.nIW++
    }
    
    s1.Tab = tab
    
    s1.Time, err = s1.IWs[0].ParseDate()
    return
}

func (s1 S1SLC) Move(dir string) (ms1 S1SLC, err error) {
    ferr := merr.Make("S1SLC.Move")
    
    newtab := filepath.Join(dir, filepath.Base(s1.Tab))
    
    var file *os.File
    if file, err = os.Create(newtab); err != nil {
        err = utils.FileOpenErr.Wrap(err, newtab)
        return
    }
    defer file.Close()
    
    for ii := 0; ii < s1.nIW; ii++ {
        if ms1.IWs[ii], err = s1.IWs[ii].Move(dir); err != nil {
            return
        }
        
        IW := ms1.IWs[ii]
        
        line := fmt.Sprintf("%s %s %s\n", IW.Dat, IW.Par, IW.TOPS_par)
        
        if _, err = file.WriteString(line); err != nil {
            err = utils.FileWriteErr.Wrap(err, newtab)
            return 
        }
    }
    
    ms1.Tab, ms1.nIW, ms1.Time = newtab, s1.nIW, s1.Time
    
    return ms1, nil
}

func (s1 S1SLC) Exist() (b bool, err error) {
    ferr := merr.Make("S1SLC.Exist")
    
    for _, iw := range s1.IWs {
        if b, err = iw.Exist(); err != nil {
            err = ferr.WrapFmt(err,
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

var mosaic = common.Gamma.Must("SLC_mosaic_S1_TOPS")

func (s1 S1SLC) Mosaic(out base.SLC, opts MosaicOpts) (err error) {
    ferr := merr.Make("S1SLC.Mosaic")
    
    opts.Looks.Default()
    
    bflg := 0
    
    if opts.BurstWindowFlag {
        bflg = 1
    }
    
    ref := "-"
    
    if len(opts.RefTab) == 0 {
        ref = opts.RefTab
    }
    
    _, err = mosaic(s1.Tab, out.Dat, out.Par, opts.Looks.Rng,
        opts.Looks.Azi, bflg, ref)
    
    if err != nil {
        return ferr.WrapFmt(err, "failed to mosaic '%s'", s1.Tab)
    }
    
    return nil
}

var derampRef = common.Gamma.Must("S1_deramp_TOPS_reference")

func (s1 S1SLC) DerampRef() (ds1 S1SLC, err error) {
    ferr := merr.Make("S1SLC.DerampRef")
    
    tab := s1.Tab
    
    if _, err = derampRef(tab); err != nil {
        err = ferr.WrapFmt(err, "failed to deramp reference S1SLC '%s'", tab)
        return
    }
    
    tab += ".deramp"
    
    if ds1, err = FromTabfile(tab); err != nil {
        err = ferr.WrapFmt(err, "failed to import S1SLC from tab '%s'", tab)
        return
    }
    
    return ds1, nil
}

var derampSlave = common.Gamma.Must("S1_deramp_TOPS_slave")

func (s1 S1SLC) DerampSlave(ref *S1SLC, looks common.RngAzi, keep bool) (ret S1SLC, err error) {
    ferr := merr.Make("S1SLC.DerampSlave")
    
    looks.Default()
    
    reftab, tab, id := ref.Tab, s1.Tab, common.DateShort.Format(s1)
    
    clean := 1
    
    if keep {
        clean = 0
    }
    
    _, err = derampSlave(tab, id, reftab, looks.Rng, looks.Azi, clean)
    
    if err != nil {
        err = ferr.WrapFmt(err,
            "failed to deramp slave S1SLC '%s', reference: '%s'", tab, reftab)
        return
    }
    
    tab += ".deramp"
    
    if ret, err = FromTabfile(tab); err != nil {
        err = ferr.WrapFmt(err, "failed to import S1SLC from tab '%s'", tab)
        return
    }
    
    return ret, nil
}

func (s1 S1SLC) RSLC(outDir string) (ret S1SLC, err error) {
    ferr := merr.Make("S1SLC.RSLC")
    
    tab := strings.ReplaceAll(filepath.Base(s1.Tab), "SLC_tab", "RSLC_tab")
    tab = filepath.Join(outDir, tab)

    file, err := os.Create(tab)

    if err != nil {
        err = utils.FileCreateErr.Wrap(err, tab)
        return
    }
    
    defer file.Close()

    for ii := 0; ii < s1.nIW; ii++ {
        IW := s1.IWs[ii]
        
        dat := filepath.Join(outDir, strings.ReplaceAll(filepath.Base(IW.Dat), "slc", "rslc"))
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
            err = utils.FileWriteErr.Wrap(err, tab)
            return
        }
    }

    ret.Tab, ret.nIW = tab, s1.nIW

    return ret, nil
}

var MLIFun = common.Gamma.SelectFun("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *S1SLC) MLI(mli *base.MLI, opt *base.MLIOpt) (err error) {
    ferr := merr.Make("S1SLC.MLI")
    opt.Parse()
    
    wflag := 0
    
    if opt.WindowFlag {
        wflag = 1
    }
    
    _, err = MLIFun(s1.Tab, mli.Dat, mli.Par, opt.Looks.Rng, opt.Looks.Azi,
                     wflag, opt.RefTab)
    
    return
}

