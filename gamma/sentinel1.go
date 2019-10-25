package gamma

import (
    "fmt"
    "log"
    "math"
    "os"
    "time"
    fp "path/filepath"
    str "strings"
)

const (
    nMaxBurst = 10
    maxIW     = 3
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

    IWInfo struct {
        nburst int
        extent Rect
        bursts [nMaxBurst]float64
    }
    
    IWInfos [maxIW]IWInfo

    S1IW struct {
        dataFile
        TOPS_par Params
    }

    IWs [maxIW]S1IW

    S1SLC struct {
        nIW int
        Tab string
        time.Time
        IWs
    }
)

var (
    burstCorner CmdFun

    burstCorners = Gamma.selectFun("ScanSAR_burst_corners",
        "SLC_burst_corners")
    calibPath = fp.Join("annotation", "calibration")

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

func NewS1Zip(zipPath, root string) (ret *S1Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"

    zipBase := fp.Base(zipPath)
    ret = &S1Zip{Path: zipPath, zipBase: zipBase}

    ret.mission = str.ToLower(zipBase[:3])
    ret.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    if ret.date, err = NewDate(DLong, start, stop); err != nil {
        err = Handle(err, "Could not create new date from strings: '%s' '%s'",
            start, stop)
        return
    }

    ret.mode = zipBase[4:6]
    safe := str.ReplaceAll(zipBase, ".zip", ".SAFE")
    tpl := fmt.Sprintf(rexTemplate, ret.mission)

    ret.Templates = [6]string{
        tiff:      fp.Join(safe, "measurement", tpl+".tiff"),
        annot:     fp.Join(safe, "annotation", tpl+".xml"),
        calib:     fp.Join(safe, calibPath, fmt.Sprintf("calibration-%s.xml", tpl)),
        noise:     fp.Join(safe, calibPath, fmt.Sprintf("noise-%s.xml", tpl)),
        preview:   fp.Join(safe, "preview", "product-preview.html"),
        quicklook: fp.Join(safe, "preview", "quick-look.png"),
    }

    ret.Safe = safe
    ret.Root = fp.Join(root, ret.Safe)
    
    for _, s1path := range S1DirPaths {
        path := fp.Join(ret.Root, s1path)
        err = os.MkdirAll(path, os.ModePerm)
        
        if err != nil {
            err = DirCreateErr.Wrap(err, path)
            return
        }
    }
    
    
    ret.productType = zipBase[7:10]
    ret.resolution = string(zipBase[10])
    ret.level = string(zipBase[12])
    ret.productClass = string(zipBase[13])
    ret.pol = zipBase[14:16]
    ret.absoluteOrbit = zipBase[49:55]
    ret.DTID = str.ToLower(zipBase[56:62])
    ret.UID = zipBase[63:67]

    return ret, nil
}

func (s1 *S1Zip) Info(exto *ExtractOpt) (ret IWInfos, err error) {
    ext, err := s1.newExtractor(exto)
    var _annot string

    if err != nil {
        err = Handle(err, "Failed to create new S1Extractor!")
        return
    }

    defer ext.Close()

    for ii := 1; ii < 4; ii++ {
        _annot, err = ext.extract(annot, ii)

        if err != nil {
            err = Handle(err, "Failed to extract annotation file from '%s'!",
                s1.Path)
            return
        }

        if ret[ii-1], err = iwInfo(_annot); err != nil {
            err = Handle(err,
                "Parsing of IW information of annotation file '%s' failed!",
                _annot)
            return
        }

    }
    return ret, nil
}

func NewIW(dat, par, TOPS_par string) (ret S1IW, err error) {
    ret.dataFile, err = NewDataFile(dat, par, Unknown)

    if err != nil {
        err = Handle(err,
            "failed to create DataFile with dat: '%s' and par '%s'",
            dat, par)
        return
    }

    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }

    ret.TOPS_par = NewGammaParam(TOPS_par)

    return ret, nil
}

const (
    ParseTabErr Werror = "failed to parse tabfile '%s'"
)

func FromTabfile(tab string) (ret S1SLC, err error) {
    fmt.Printf("Parsing tabfile: '%s'.\n", tab)
    
    file, err := NewReader(tab)
    
    if err != nil {
        err = FileOpenErr.Wrap(err, tab)
        return
    }
    
    defer file.Close()

    ret.nIW = 0
    
    for file.Scan() {
        line := file.Text()
        split := str.Split(line, " ")
        
        log.Printf("Parsing IW%d\n", ret.nIW + 1)
        
        ret.IWs[ret.nIW], err = NewIW(split[0], split[1], split[2])
        
        if err != nil {
            err = Handle(err, "failed to parse IW files from line '%s'", line)
            return
        }
        
        ret.nIW++
    }
    
    ret.Tab = tab
        
    ret.Time, err = ret.IWs[0].Date()
    
    if err != nil {
        err = Handle(err, "failed to retreive date for '%s'", tab)
        return
    }
    
    return ret, nil
}

func (s1 S1SLC) TypeStr() string {
    return "S1SLC"
}

func (s1 *S1SLC) Exist() (ret bool, err error) {
    var exist bool
    for _, iw := range s1.IWs {
        exist, err = iw.Exist()

        if err != nil {
            err = fmt.Errorf("Could not determine whether IW datafile exists!")
            return
        }

        if !exist {
            return false, nil
        }
    }
    return true, nil
}

func (iw *S1IW) Move(dir string) error {
    slc, par, TOPS_par := iw.dataFile.Dat, iw.dataFile.Params.Par, iw.TOPS_par.Par
    
    dst := fp.Join(dir, fp.Base(slc))
    err := os.Rename(slc, dst)

    if err != nil {
        return MoveErr.Wrap(err, slc, dst)
        //return Handle(err, "failed to move file '%s' to '%s'", slc, dst)
    }
    
    iw.dataFile.Dat = dst
    
    dst = fp.Join(dir, fp.Base(par))
    err = os.Rename(par, dst)

    if err != nil {
        return MoveErr.Wrap(err, par, dst)
    }
    
    iw.dataFile.Params.Par = dst
    
    dst = fp.Join(dir, fp.Base(TOPS_par))
    err = os.Rename(TOPS_par, dst)

    if err != nil {
        return MoveErr.Wrap(err, TOPS_par, dst)
    }
    
    iw.TOPS_par.Par = dst
    
    return nil
}

func (s1 *S1SLC) Move(dir string) error {
    newtab := fp.Join(dir, fp.Base(s1.Tab))
    
    file, err := os.Create(newtab)
    
    if err != nil {
        return FileOpenErr.Wrap(err, newtab)
    }
    
    defer file.Close()
    
    for ii := 0; ii < s1.nIW; ii++ {
        IW := &s1.IWs[ii]
        
        err := IW.Move(dir)
        
        if err != nil {
            return err
        }
        
        line := fmt.Sprintf("%s %s %s\n", IW.Dat, IW.Par, IW.TOPS_par.Par)
        
        _, err = file.WriteString(line)
        
        if err != nil {
            return FileWriteErr.Wrap(err, newtab)
        }
    }
    
    s1.Tab = newtab
    return nil
}

type MosaicOpts struct {
    Looks RngAzi
    BurstWindowFlag bool
    RefTab string
}

var mosaic = Gamma.Must("SLC_mosaic_S1_TOPS")

func (s1 *S1SLC) Mosaic(opts MosaicOpts) (ret SLC, err error) {
    opts.Looks.Default()
    
    bflg := 0
    
    if opts.BurstWindowFlag {
        bflg = 1
    }
    
    ref := "-"
    
    if len(opts.RefTab) == 0 {
        ref = opts.RefTab
    }
    
    tmp := ""
    
    if tmp, err = TmpFileExt("slc"); err != nil {
        //err = Handle(err, "failed to create tmp file")
        return ret, err
    }
    
    if ret, err = NewSLC(tmp, ""); err != nil {
        err = DataCreateErr.Wrap(err, "SLC")
        return
    }
    
    _, err = mosaic(s1.Tab, ret.Dat, ret.Par, opts.Looks.Rng, opts.Looks.Azi,
                    bflg, ref)
    
    if err != nil {
        err = Handle(err, "failed to mosaic '%s'", s1.Tab)
        return
    }
    
    return ret, nil
}

var derampRef = Gamma.Must("S1_deramp_TOPS_reference")

func (s1 *S1SLC) DerampRef() (ret S1SLC, err error) {
    tab := s1.Tab
    
    if _, err = derampRef(tab); err != nil {
        err = Handle(err, "failed to deramp reference S1SLC '%s'", tab)
        return
    }
    
    tab += ".deramp"
    
    if ret, err = FromTabfile(tab); err != nil {
        err = Handle(err, "failed to import S1SLC from tab '%s'", tab)
        return
    }
    return ret, nil
}

var derampSlave = Gamma.Must("S1_deramp_TOPS_slave")

func (s1 *S1SLC) DerampSlave(ref *S1SLC, looks RngAzi, keep bool) (ret S1SLC, err error) {
    looks.Default()
    
    reftab, tab, id := ref.Tab, s1.Tab, s1.Format(DateShort)
    
    clean := 1
    
    if keep {
        clean = 0
    }
    
    _, err = derampSlave(tab, id, reftab, looks.Rng, looks.Azi, clean)
    
    if err != nil {
        err = Handle(err, "failed to deramp slave S1SLC '%s', reference: '%s'",
            tab, reftab)
        return
    }
    
    tab += ".deramp"
    
    if ret, err = FromTabfile(tab); err != nil {
        err = Handle(err, "failed to import S1SLC from tab '%s'", tab)
        return
    }
    return ret, nil
}

func makePoint(info Params, max bool) (ret Point, err error) {
    var tpl_lon, tpl_lat string

    if max {
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    } else {
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if ret.X, err = info.Float(tpl_lon, 0); err != nil {
        err = Handle(err, "Could not get Longitude value!")
        return
    }

    if ret.Y, err = info.Float(tpl_lat, 0); err != nil {
        err = Handle(err, "Could not get Latitude value!")
        return
    }

    return ret, nil
}

func iwInfo(path string) (ret IWInfo, err error) {
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);

    if len(path) == 0 {
        err = Handle(nil, "path to annotation file is an empty string: '%s'",
            path)
        return
    }

    // Check(err, "Failed to retreive IW number from %s", path);

    par, err := TmpFile()

    if err != nil {
        return ret, err
    }

    TOPS_par, err := TmpFile()

    if err != nil {
        return ret, err
    }

    _, err = Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par)

    if err != nil {
        err = Handle(err, "failed to import parameter files from '%s'",
            path)
        return
    }

    _info, err := burstCorners(par, TOPS_par)

    if err != nil {
        err = Handle(err, "failed to parse parameter files")
        return
    }

    // TODO: generic reader Params
    info := FromString(_info, ":")
    TOPS := NewGammaParam(TOPS_par)

    nburst, err := TOPS.Int("number_of_bursts", 0)

    if err != nil {
        err = Handle(err, "failed to retreive number of bursts")
        return
    }

    var numbers [nMaxBurst]float64

    for ii := 1; ii < nburst+1; ii++ {
        tpl := fmt.Sprintf(burstTpl, ii)

        numbers[ii-1], err = TOPS.Float(tpl, 0)

        if err != nil {
            err = Handle(err, "failed to get burst number: '%s'", tpl)
            return
        }
    }

    max, err := makePoint(info, true)

    if err != nil {
        err = Handle(err, "failed to create max latlon point")
        return
    }

    min, err := makePoint(info, false)

    if err != nil {
        err = Handle(err, "failed to create min latlon point")
        return
    }

    return IWInfo{nburst: nburst, extent: Rect{Min: min, Max: max},
        bursts: numbers}, nil
}

func (p *Point) inIWs(IWs IWInfos) bool {
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

    return int(dburst + 1.0 + (dburst/(0.001+diffSqrt))*0.5)
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
    sum := 0.0

    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            return 0.0, fmt.Errorf(
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
    path := fp.Join(s1.Root, mode)
    
    dat = fp.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    par = dat + ".par"
    
    return
}

func (s1 *S1Zip) SLCNames(mode, pol string, ii int) (dat, par, TOPS string) {
    slcPath := fp.Join(s1.Root, mode)

    dat = fp.Join(slcPath, fmt.Sprintf("iw%d_%s.%s", ii, pol, mode))
    par = dat + ".par"
    TOPS = dat + ".TOPS_par"

    return
}

func (s1 *S1Zip) SLC(pol string) (ret S1SLC, err error) {
    const mode = "slc"
    tab := s1.tabName(mode, pol)

    var exist bool
    exist, err = Exist(tab)

    if err != nil {
        return
    }

    if !exist {
        err = Handle(err, "tabfile '%s' does not exist", tab)
        return
    }

    for ii := 1; ii < 4; ii++ {
        dat, par, TOPS_par := s1.SLCNames(mode, pol, ii)
        ret.IWs[ii-1], err = NewIW(dat, par, TOPS_par)

        if err != nil {
            err = DataCreateErr.Wrap(err, "IW")
            return
        }
    }

    ret.Tab, ret.nIW = tab, 3

    return ret, nil
}

func (s1 *S1SLC) RSLC(outDir string) (ret S1SLC, err error) {
    tab := fp.Join(outDir, str.ReplaceAll(fp.Base(s1.Tab), "SLC_tab", "RSLC_tab"))

    file, err := os.Create(tab)

    if err != nil {
        err = FileCreateErr.Wrap(err, tab)
        return
    }
    
    defer file.Close()

    for ii := 0; ii < s1.nIW; ii++ {
        IW := s1.IWs[ii]
        
        dat := fp.Join(outDir, str.ReplaceAll(fp.Base(IW.Dat), "slc", "rslc"))
        par := dat + ".par"
        TOPS_par := dat + ".TOPS_par"
        
        ret.IWs[ii], err = NewIW(dat, par, TOPS_par)

        if err != nil {
            err = DataCreateErr.Wrap(err, "IW")
            //err = Handle(err, "failed to create new IW")
            return
        }
        
        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            err = FileWriteErr.Wrap(err, tab)
            return
        }
    }

    ret.Tab, ret.nIW = tab, s1.nIW

    return ret, nil
}

var MLIFun = Gamma.selectFun("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *S1SLC) MLI(mli *MLI, opt *MLIOpt) error {
    opt.Parse()
    
    wflag := 0
    
    if opt.windowFlag {
        wflag = 1
    }
    
    _, err := MLIFun(s1.Tab, mli.Dat, mli.Par, opt.Looks.Rng, opt.Looks.Azi,
                     wflag, opt.refTab)
    
    if err != nil {
        return DataCreateErr.Wrap(err, "MLI")
        //return Handle(err, "failed to create MLI file '%s'", mli.Dat)
    }
    
    return nil
}

func (s1 *S1Zip) MLI(mode, pol string, opt *MLIOpt) (ret MLI, err error) {
    path := fp.Join(s1.Root, mode)

    dat := fp.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    par := dat + ".par"
    
    ret, err = NewMLI(dat, par)
    
    if err != nil {
        err = DataCreateErr.Wrap(err, "MLI")
        return
    }
    
    exist, err := ret.Exist()
    
    if err != nil {
        //err = FileExistErr.Wrap(err, 
        err = Handle(err, "failed to check whether MLI exists")
        return
    }
    
    if exist {
        return ret, nil
    }
    
    slc, err := s1.SLC(pol)
    
    if err != nil {
        err = DataCreateErr.Wrap(err, "S1SLC")
        return
    }
    
    err = slc.MLI(&ret, opt)
    
    if err != nil {
        //err = Handle(err, "failed to check create MLI file")
        return ret, err
    }
    
    return ret, nil
}

func (s1 *S1Zip) tabName(mode, pol string) string {
    return fp.Join(s1.Root, mode, fmt.Sprintf("%s.tab", pol))
}

const (
    ExtractErr Werror = "failed to extract %s file from '%s'"
)

func (s1 *S1Zip) ImportSLC(exto *ExtractOpt) (err error) {
    var _annot, _calib, _tiff, _noise string
    ext, err := s1.newExtractor(exto)

    if err != nil {
        err = Handle(err, "failed to create S1Extractor")
        return
    }

    defer ext.Close()

    path, pol := s1.Path, exto.pol
    tab := s1.tabName("slc", pol)

    file, err := os.Create(tab)

    if err != nil {
        err = FileOpenErr.Wrap(err, tab)
        return
    }

    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        if _annot, err = ext.extract(annot, ii); err != nil {
            err = ExtractErr.Wrap(err, "annotation", path)
            return
        }

        if _calib, err = ext.extract(calib, ii); err != nil {
            err = ExtractErr.Wrap(err, "calibration", path)
            return
        }

        if _tiff, err = ext.extract(tiff, ii); err != nil {
            err = ExtractErr.Wrap(err, "TIFF", path)
            return
        }

        if _noise, err = ext.extract(noise, ii); err != nil {
            err = ExtractErr.Wrap(err, "noise", path)
            return
        }

        dat, par, TOPS_par := s1.SLCNames("slc", pol, ii)

        _, err = Gamma["par_S1_SLC"](_tiff, _annot, _calib, _noise, par, dat,
            TOPS_par)
        if err != nil {
            err = Handle(err, "failed to import datafiles into gamma format")
            return
        }

        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            err = FileWriteErr.Wrap(err, tab)
            return
        }
    }

    return nil
}

func (s1 *S1Zip) Quicklook(exto *ExtractOpt) (ret string, err error) {
    ext, err := s1.newExtractor(exto)

    if err != nil {
        err = Handle(err, "failed to create new S1Extractor")
        return
    }

    defer ext.Close()

    path := s1.Path

    ret, err = ext.extract(quicklook, 0)

    if err != nil {
        err = ExtractErr.Wrap(err, "annotation", path)
        return
    }

    return ret, nil
}

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d ByDate) Less(i, j int) bool {
    return Before(d[i], d[j])
}
