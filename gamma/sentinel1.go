package gamma

import (
    "fmt"
    "log"
    "math"
    "os"
    fp "path/filepath"
    str "strings"
)

const (
    nMaxBurst = 10
    nIW       = 3
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
        date          `json:"date"`
    }
    
    
    S1Zips []*S1Zip
    ByDate S1Zips

    IWInfo struct {
        nburst int
        extent Rect
        bursts [nMaxBurst]float64
    }
    
    IWInfos [nIW]IWInfo

    S1IW struct {
        dataFile
        TOPS_par Params
    }

    IWs [nIW]S1IW

    S1SLC struct {
        date
        IWs IWs
        tab string
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
)

func init() {
    var ok bool

    if burstCorner, ok = Gamma["ScanSAR_burst_corners"]; !ok {
        if burstCorner, ok = Gamma["SLC_burst_corners"]; !ok {
            log.Fatalf("No Fun.")
        }
    }
}

func NewS1Zip(zipPath, root string) (ret *S1Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"

    zipBase := fp.Base(zipPath)
    ret = &S1Zip{Path: zipPath, zipBase: zipBase}

    ret.mission = str.ToLower(zipBase[:3])
    ret.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    if ret.date, err = NewDate(long, start, stop); err != nil {
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
            err = Handle(err, "Failed to create directory '%s'!", path)
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
    ret.dataFile, err = NewDataFile(dat, par, "par")

    if err != nil {
        err = Handle(err,
            "Failed to create DataFile with dat: '%s' and par '%s'!",
            dat, par)
        return
    }

    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }

    ret.TOPS_par = NewGammaParam(TOPS_par)
    ret.files = []string{dat, par, TOPS_par}

    return ret, nil
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

func makePoint(info Params, max bool) (ret Point, err error) {
    var tpl_lon, tpl_lat string

    if max {
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    } else {
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if ret.X, err = info.Float(tpl_lon); err != nil {
        err = Handle(err, "Could not get Longitude value!")
        return
    }

    if ret.Y, err = info.Float(tpl_lat); err != nil {
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
        err = Handle(err, "Failed to create tmp file!")
        return
    }

    TOPS_par, err := TmpFile()

    if err != nil {
        err = Handle(err, "Failed to create tmp file!")
        return
    }

    _, err = Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par)

    if err != nil {
        err = Handle(err, "Could not import parameter files from '%s'!",
            path)
        return
    }

    _info, err := burstCorners(par, TOPS_par)

    if err != nil {
        err = Handle(err, "Failed to parse parameter files!")
        return
    }

    // TODO: generic reader Params
    info := FromString(_info, ":")
    TOPS := NewGammaParam(TOPS_par)

    nburst, err := TOPS.Int("number_of_bursts")

    if err != nil {
        err = Handle(err, "Could not retreive number of bursts!")
        return
    }

    var numbers [nMaxBurst]float64

    for ii := 1; ii < nburst+1; ii++ {
        tpl := fmt.Sprintf(burstTpl, ii)

        numbers[ii-1], err = TOPS.Float(tpl)

        if err != nil {
            err = Handle(err, "Could not get burst number: '%s'", tpl)
            return
        }
    }

    max, err := makePoint(info, true)

    if err != nil {
        err = Handle(err, "Could not create max latlon point!")
        return
    }

    min, err := makePoint(info, false)

    if err != nil {
        err = Handle(err, "Could not create min latlon point!")
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
                "In: IWInfos.AbsDiff: number of burst in first SLC IW%d (%d) "+
                    "is not equal to the number of burst in the second SLC IW%d (%d)",
                ii+1, nburst1, ii+1, nburst2)
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
        err = Handle(err, "tabfile '%s' does not exist!", tab)
        return
    }

    for ii := 1; ii < 4; ii++ {
        dat, par, TOPS_par := s1.SLCNames(mode, pol, ii)
        ret.IWs[ii-1], err = NewIW(dat, par, TOPS_par)

        if err != nil {
            err = Handle(err, "Could not create new IW!")
            return
        }
    }

    ret.tab = tab

    return ret, nil
}

func (s1 *S1Zip) RSLC(pol string) (ret S1SLC, err error) {
    const mode = "rslc"
    tab := s1.tabName(mode, pol)

    file, err := os.Create(tab)

    if err != nil {
        err = Handle(err, "Failed to create file '%s'!", tab)
        return
    }
    
    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        dat, par, TOPS_par := s1.SLCNames(mode, pol, ii)
        ret.IWs[ii-1], err = NewIW(dat, par, TOPS_par)

        if err != nil {
            err = Handle(err, "Could not create new IW!")
            return
        }
        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            err = Handle(err, "Failed to write line '%s' to file'%s'!",
                line, tab)
            return
        }
    }

    ret.tab = tab

    return ret, nil
}

var MLIFun = Gamma.selectFun("multi_look_ScanSAR", "multi_S1_TOPS")

func (s1 *S1SLC) MLI(mli *MLI, opt *MLIOpt) error {
    opt.Parse()
    
    wflag := 0
    
    if opt.windowFlag {
        wflag = 1
    }
    
    _, err := MLIFun(s1.tab, mli.Dat, mli.Par, opt.Looks.Rng, opt.Looks.Azi,
                     wflag, opt.refTab)
    
    if err != nil {
        return Handle(err, "failed to create MLI file '%s'", mli.Dat)
    }
    
    return nil
}

func (s1 *S1Zip) MLI(mode, pol string, opt *MLIOpt) (ret MLI, err error) {
    path := fp.Join(s1.Root, mode)

    dat := fp.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    par := dat + ".par"
    
    ret, err = NewMLI(dat, par)
    
    if err != nil {
        err = Handle(err, "failed to create MLI struct")
        return
    }
    
    exist, err := ret.Exist()
    
    if err != nil {
        err = Handle(err, "failed to check whether MLI exists")
        return
    }
    
    if exist {
        return ret, nil
    }
    
    slc, err := s1.SLC(pol)
    
    if err != nil {
        err = Handle(err, "failed to create S1SLC struct")
        return
    }
    
    err = slc.MLI(&ret, opt)
    
    if err != nil {
        err = Handle(err, "failed to check create MLI file")
        return
    }
    
    return ret, nil
}

func (s1 *S1Zip) tabName(mode, pol string) string {
    return fp.Join(s1.Root, mode, fmt.Sprintf("%s.tab", pol))
}

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
        err = Handle(err, "failed to open file '%s'", tab)
        return
    }

    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        _annot, err = ext.extract(annot, ii)

        if err != nil {
            err = Handle(err, "failed to extract annotation file from '%s'",
                path)
            return
        }

        _calib, err = ext.extract(calib, ii)

        if err != nil {
            err = Handle(err, "failed to extract calibration file from '%s'",
                path)
            return
        }

        _tiff, err = ext.extract(tiff, ii)

        if err != nil {
            err = Handle(err, "failed to extract tiff file from '%s'",
                path)
            return
        }

        _noise, err = ext.extract(noise, ii)

        if err != nil {
            err = Handle(err, "failed to extract noise file from '%s'",
                path)
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
            err = Handle(err, "failed to write line '%s' to file'%s'",
                line, tab)
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
        err = Handle(err, "failed to extract annotation file from '%s'",
            path)
        return
    }

    return ret, nil
}

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d ByDate) Less(i, j int) bool {
    return Before(d[i], d[j])
}
