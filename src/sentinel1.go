package gamma

import (
    "fmt"
    "log"
    "math"
    "os"
    "time"
    "path/filepath"
    "strings"
)

const (
    nMaxBurst = 10
    burstTpl  = "burst_asc_node_%d"
    nTemplate = 6
)

const (
    tiff tplType = iota
    annot
    calib
    noise
    preview
    quicklook
)


type (
    tplType   int
    templates [nTemplate]string
    
    S1Zip struct {
        Path          string    `json:"path"`
        Root          string    `json:"root"`
        zipBase       string
        mission       string
        dateStr       string
        mode          string
        productType   string
        resolution    string
        Safe          string    `json:"safe"`
        level         string
        productClass  string
        pol           string
        absoluteOrbit string
        DTID          string    `json:"-"`
        UID           string    `json:"-"`
        Templates     templates `json:"templates"`
        date                    `json:"date"`
    }
    
    
    S1Zips []*S1Zip
    ByDate S1Zips
)

var (
    burstCorner CmdFun

    burstCorners = Gamma.selectFun("ScanSAR_burst_corners",
        "SLC_burst_corners")
    calibPath = filepath.Join("annotation", "calibration")

    fmtNeeded = [nTemplate]bool{
        tiff:      true,
        annot:     true,
        calib:     true,
        noise:     true,
        preview:   false,
        quicklook: false,
    }
    
    S1DirPaths = [4]string{"slc", "rslc", "mli", "rmli"}
    S1SLCType = "S1SLC"
)

func NewS1Zip(zipPath, pol string) (s1 *S1Zip, err error) {
    ferr := merr.Make("NewS1Zip")
    
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"

    zipBase := filepath.Base(zipPath)
    s1 = &S1Zip{Path: zipPath, zipBase: zipBase, pol: pol}

    s1.mission = strings.ToLower(zipBase[:3])
    s1.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    if s1.date, err = NewDate(DLong, start, stop); err != nil {
        err = ferr.WrapFmt(err,
            "Could not create new date from strings: '%s' '%s'", start, stop)
        return
    }

    s1.mode = zipBase[4:6]
    safe := strings.ReplaceAll(zipBase, ".zip", ".SAFE")
    tpl := fmt.Sprintf(rexTemplate, s1.mission)

    s1.Templates = templates{
        tiff:      filepath.Join(safe, "measurement", tpl+".tiff"),
        annot:     filepath.Join(safe, "annotation", tpl+".xml"),
        calib:     filepath.Join(safe, calibPath, fmt.Sprintf("calibration-%s.xml", tpl)),
        noise:     filepath.Join(safe, calibPath, fmt.Sprintf("noise-%s.xml", tpl)),
        preview:   filepath.Join(safe, "preview", "product-preview.html"),
        quicklook: filepath.Join(safe, "preview", "quick-look.png"),
    }

    s1.Safe = safe
    s1.productType = zipBase[7:10]
    s1.resolution = string(zipBase[10])
    s1.level = string(zipBase[12])
    s1.productClass = string(zipBase[13])
    s1.pol = zipBase[14:16]
    s1.absoluteOrbit = zipBase[49:55]
    s1.DTID = strings.ToLower(zipBase[56:62])
    s1.UID = zipBase[63:67]

    return s1, nil
}

func (s1 *S1Zip) Info(dst string) (iws IWInfos, err error) {
    ferr := merr.Make("S1Zip.Info")
    
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer ext.Close()

    var _annot string
    for ii := 1; ii < 4; ii++ {
        _annot = ext.Extract(annot, ii)
        
        if err = ext.Wrap(); err != nil {
            err = ferr.Wrap(err)
            return
        }

        if iws[ii-1], err = iwInfo(_annot); err != nil {
            err = ferr.WrapFmt(err,
                "Parsing of IW information of annotation file '%s' failed!",
                _annot)
            return
        }

    }
    return iws, nil
}

const maxIW = 3

type(  
    S1IW struct {
        DatParFile
        TOPS_par Params
    }

    IWs [maxIW]S1IW
)

func NewIW(dat, par, TOPS_par string) (iw S1IW) {
    iw.Dat = dat
    
    if len(par) == 0 {
        par = dat + ".par"
    }
    
    iw.Params = Params{Par: par, Sep: ":"}
    
    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }

    iw.TOPS_par = Params{Par: TOPS_par, Sep: ":"}

    return
}

func (iw S1IW) Move(dir string) (miw S1IW, err error) {
    ferr := merr.Make("S1IW.Move")
    
    if miw.DatParFile, err = iw.DatParFile.Move(dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    miw.TOPS_par.Par, err = Move(iw.TOPS_par.Par, dir)
    
    return miw, nil
}


const (
    ParseTabErr Werror = "failed to parse tabfile '%s'"
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
    
    var file Reader
    if file, err = NewReader(tab); err != nil {
        err = ferr.Wrap(FileOpenErr.Wrap(err, tab))
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
    
    if s1.Time, err = s1.IWs[0].ParseDate(); err != nil {
        err = ferr.WrapFmt(err, "failed to retreive date for '%s'", tab)
    }
    
    return
}


func (s1 S1SLC) jsonName() string {
    return s1.Tab + ".json"
}

func (s1 S1SLC) Move(dir string) (ms1 S1SLC, err error) {
    ferr := merr.Make("S1SLC.Move")
    
    newtab := filepath.Join(dir, filepath.Base(s1.Tab))
    
    var file *os.File
    if file, err = os.Create(newtab); err != nil {
        err = ferr.Wrap(FileOpenErr.Wrap(err, newtab))
        return
    }
    defer file.Close()
    
    for ii := 0; ii < s1.nIW; ii++ {
        if ms1.IWs[ii], err = s1.IWs[ii].Move(dir); err != nil {
            return
        }
        
        IW := ms1.IWs[ii]
        
        line := fmt.Sprintf("%s %s %s\n", IW.Dat, IW.Par, IW.TOPS_par.Par)
        
        if _, err = file.WriteString(line); err != nil {
            err = ferr.Wrap(FileWriteErr.Wrap(err, newtab))
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
    Looks RngAzi
    BurstWindowFlag bool
    RefTab string
}

var mosaic = Gamma.Must("SLC_mosaic_S1_TOPS")

func (s1 S1SLC) Mosaic(opts MosaicOpts) (slc SLC, err error) {
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
    
    if slc, err = TmpSLC(); err != nil {
        err = ferr.Wrap(err)
        return
    }

    _, err = mosaic(s1.Tab, slc.Dat, slc.Par, opts.Looks.Rng,
        opts.Looks.Azi, bflg, ref)
    
    if err != nil {
        err = ferr.WrapFmt(err, "failed to mosaic '%s'", s1.Tab)
        return
    }
    
    return slc, nil
}

var derampRef = Gamma.Must("S1_deramp_TOPS_reference")

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

var derampSlave = Gamma.Must("S1_deramp_TOPS_slave")

func (s1 S1SLC) DerampSlave(ref *S1SLC, looks RngAzi, keep bool) (ret S1SLC, err error) {
    ferr := merr.Make("S1SLC.DerampSlave")
    
    looks.Default()
    
    reftab, tab, id := ref.Tab, s1.Tab, s1.Format(DateShort)
    
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
        err = ferr.Wrap(FileCreateErr.Wrap(err, tab))
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
            err = ferr.Wrap(FileWriteErr.Wrap(err, tab))
            return
        }
    }

    ret.Tab, ret.nIW = tab, s1.nIW

    return ret, nil
}

var MLIFun = Gamma.selectFun("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *S1SLC) MLI(mli *MLI, opt *MLIOpt) error {
    ferr := merr.Make("S1SLC.MLI")
    opt.Parse()
    
    wflag := 0
    
    if opt.windowFlag {
        wflag = 1
    }
    
    _, err := MLIFun(s1.Tab, mli.Dat, mli.Par, opt.Looks.Rng, opt.Looks.Azi,
                     wflag, opt.refTab)
    
    if err != nil {
        return ferr.Wrap(StructCreateError.Wrap(err, "MLI"))
    }
    
    return nil
}

func makePoint(info Params, max bool) (ret Point, err error) {
    ferr := merr.Make("makePoint")
    var tpl_lon, tpl_lat string

    if max {
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    } else {
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if ret.X, err = info.Float(tpl_lon, 0); err != nil {
        err = ferr.WrapFmt(err, "Could not get Longitude value!")
        return
    }

    if ret.Y, err = info.Float(tpl_lat, 0); err != nil {
        err = ferr.WrapFmt(err, "Could not get Latitude value!")
        return
    }

    return ret, nil
}

type(
    IWInfo struct {
        nburst int
        extent Rect
        bursts [nMaxBurst]float64
    }
    
    IWInfos [maxIW]IWInfo
)

var parCmd = Gamma.Must("par_S1_SLC")

func iwInfo(path string) (ret IWInfo, err error) {
    ferr := merr.Make("iwInfo")
    
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);

    if len(path) == 0 {
        err = ferr.Fmt("path to annotation file is an empty string: '%s'",
            path)
        return
    }

    // Check(err, "Failed to retreive IW number from %s", path);

    par, err := TmpFile("")

    if err != nil {
        return ret, ferr.Wrap(err)
    }

    TOPS_par := par + ".TOPS_par"

    _, err = parCmd(nil, path, nil, nil, par, nil, TOPS_par)

    if err != nil {
        err = ferr.WrapFmt(err, "failed to import parameter files from '%s'",
            path)
        return
    }

    _info, err := burstCorners(par, TOPS_par)

    if err != nil {
        err = ferr.WrapFmt(err, "failed to parse parameter files")
        return
    }

    // TODO: generic reader Params
    info := FromString(_info, ":")
    TOPS := NewGammaParam(TOPS_par)

    nburst, err := TOPS.Int("number_of_bursts", 0)

    if err != nil {
        err = ferr.WrapFmt(err, "failed to retreive number of bursts")
        return
    }

    var numbers [nMaxBurst]float64

    for ii := 1; ii < nburst+1; ii++ {
        tpl := fmt.Sprintf(burstTpl, ii)

        numbers[ii-1], err = TOPS.Float(tpl, 0)

        if err != nil {
            err = ferr.WrapFmt(err, "failed to get burst number: '%s'", tpl)
            return
        }
    }

    max, err := makePoint(info, true)

    if err != nil {
        err = ferr.WrapFmt(err, "failed to create max latlon point")
        return
    }

    min, err := makePoint(info, false)

    if err != nil {
        err = ferr.WrapFmt(err, "failed to create min latlon point")
        return
    }

    return IWInfo{nburst: nburst, extent: Rect{Min: min, Max: max},
        bursts: numbers}, nil
}

func (p Point) inIWs(IWs IWInfos) bool {
    for _, iw := range IWs {
        if p.InRect(&iw.extent) {
            return true
        }
    }
    return false
}

func (iw IWInfos) contains(aoi AOI) bool {
    sum := 0

    for _, point := range aoi {
        if point.inIWs(iw) {
            sum++
        }
    }
    return sum == 4
}

func diffBurstNum(burst1, burst2 float64) int {
    dburst := burst1 - burst2
    diffSqrt := math.Sqrt(dburst)

    return int(dburst + 1.0 + (dburst / (0.001 + diffSqrt)) * 0.5)
}

func checkBurstNum(one, two IWInfos) bool {
    for ii := 0; ii < 3; ii++ {
        if one[ii].nburst != two[ii].nburst {
            return true
        }
    }
    return false
}

func IWAbsDiff(one, two IWInfos) (float64, error) {
    ferr := merr.Make("IWAbsDiff")
    sum := 0.0

    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            return 0.0, ferr.Fmt(
                "number of burst in first SLC IW%d (%d) is not equal to " + 
                "the number of burst in the second SLC IW%d (%d)",
                ii + 1, nburst1, ii + 1, nburst2)
        }

        for jj := 0; jj < nburst1; jj++ {
            dburst := one[ii].bursts[jj] - two[ii].bursts[jj]
            sum += dburst * dburst
        }
    }

    return math.Sqrt(sum), nil
}

func (s1 *S1Zip) Names(mode, pol string) (dat, par string) {
    path := filepath.Join(s1.Root, mode)
    
    dat = filepath.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    par = dat + ".par"
    
    return
}

func (s1 *S1Zip) SLCNames(mode, pol string, ii int) (dat, par, TOPS string) {
    slcPath := filepath.Join(s1.Root, mode)

    dat = filepath.Join(slcPath, fmt.Sprintf("iw%d_%s.%s", ii, pol, mode))
    par = dat + ".par"
    TOPS = dat + ".TOPS_par"

    return
}

func (s1 *S1Zip) SLC(pol string) (ret S1SLC, err error) {
    ferr := merr.Make("S1Zip.SLC")
    
    const mode = "slc"
    tab := s1.tabName(mode, pol)

    var exist bool
    exist, err = Exist(tab)

    if err != nil {
        err = ferr.Wrap(err)
        return
    }

    if !exist {
        err = ferr.WrapFmt(err, "tabfile '%s' does not exist", tab)
        return
    }

    for ii := 1; ii < 4; ii++ {
        dat, par, TOPS_par := s1.SLCNames(mode, pol, ii)
        ret.IWs[ii-1] = NewIW(dat, par, TOPS_par)
    }

    ret.Tab, ret.nIW = tab, 3

    return ret, nil
}

func (s1 *S1Zip) MLI(mode, pol string, opt *MLIOpt) (ret MLI, err error) {
    ferr := merr.Make("S1Zip.MLI")
    //path := filepath.Join(s1.Root, mode)
    
    if ret, err = TmpMLI(); err != nil {
        err = ferr.Wrap(err)
        return
    }

    //dat := fp.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    //par := dat + ".par"
    
    //ret, err = NewMLI(dat, par)
    
    //if err != nil {
        //err = DataCreateErr.Wrap(err, "MLI")
        //return
    //}
    
    //exist, err := ret.Exist()
    
    //if err != nil {
        //err = FileExistErr.Wrap(err, 
        //err = Handle(err, "failed to check whether MLI exists")
        //return
    //}
    
    //if exist {
        //return ret, nil
    //}
    
    slc, err := s1.SLC(pol)
    
    if err != nil {
        err = StructCreateError.Wrap(err, "S1SLC")
        return
    }
    
    err = slc.MLI(&ret, opt)
    
    if err != nil {
        return ret, ferr.Wrap(err)
    }
    
    return ret, nil
}

func (s1 *S1Zip) tabName(mode, pol string) string {
    return filepath.Join(s1.Root, mode, fmt.Sprintf("%s.tab", pol))
}

const (
    ExtractErr Werror = "failed to extract %s file from '%s'"
)

func (s1 *S1Zip) ImportSLC(dst string) (err error) {
    ferr := merr.Make("S1Zip.ImportSLC")
    
    var _annot, _calib, _tiff, _noise string
    
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        return
    }

    defer ext.Close()

    pol := s1.pol
    tab := s1.tabName("slc", pol)

    file, err := os.Create(tab)
    if err != nil {
        err = ferr.Wrap(FileOpenErr.Wrap(err, tab))
        return
    }
    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        _annot = ext.Extract(annot, ii)
        _calib = ext.Extract(calib, ii)
        _tiff  = ext.Extract(tiff, ii)
        _noise = ext.Extract(noise, ii)

        if err = ext.Wrap(); err != nil {
            return ferr.Wrap(err)
        }

        dat, par, TOPS_par := s1.SLCNames("slc", pol, ii)

        _, err = Gamma["par_S1_SLC"](_tiff, _annot, _calib, _noise, par, dat,
            TOPS_par)
        if err != nil {
            err = ferr.WrapFmt(err,
                "failed to import datafiles into gamma format")
            return
        }

        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            err = ferr.Wrap(FileWriteErr.Wrap(err, tab))
            return
        }
    }

    return nil
}

func (s1 *S1Zip) Quicklook(dst string) (s string, err error) {
    ferr := merr.Make("S1Zip.Quicklook")
    
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer ext.Close()

    s = ext.Extract(quicklook, 0)
    
    if err = ext.Wrap(); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    return s, nil
}

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d ByDate) Less(i, j int) bool {
    return Before(d[i], d[j])
}
