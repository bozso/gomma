package sentinel1

import (
    "fmt"
    "math"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/date"
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/utils/params"
    "github.com/bozso/gamma/base"

    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/cli/stream"
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
    
    Zip struct {
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
        date          DateRange `json:"date"`
    }
    
    
    Zips []*Zip
    ByDate Zips
)

var (
    burstCorners = common.Select("ScanSAR_burst_corners",
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

func NewZip(zipPath, pol string) (s1 *Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"

    zipBase := filepath.Base(zipPath)
    s1.Path, s1.zipBase, s1.pol = zipPath, zipBase, pol

    s1.mission = strings.ToLower(zipBase[:3])
    s1.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    s1.date, err = NewDate(date.Long, start, stop)
    
    if err != nil {
        err = utils.WrapFmt(err,
            "Could not create new date from strings: '%s' '%s'", start, stop)
        return
    }

    s1.mode = zipBase[4:6]
    safe := strings.ReplaceAll(zipBase, ".zip", ".SAFE")
    tpl := fmt.Sprintf(rexTemplate, s1.mission)

    s1.Templates = templates{
        tiff:      filepath.Join(safe, "measurement", tpl + ".tiff"),
        annot:     filepath.Join(safe, "annotation", tpl + ".xml"),
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

func (s1 Zip) Info(dst string) (iws IWInfos, err error) {
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        return
    }
    defer ext.Close()

    var _annot string
    for ii := 1; ii < 4; ii++ {
        _annot = ext.Extract(annot, ii)
        
        if err = ext.Wrap(); err != nil {
            return
        }

        if iws[ii-1], err = iwInfo(_annot); err != nil {
            err = utils.WrapFmt(err,
                "Parsing of IW information of annotation file '%s' failed!",
                _annot)
            return
        }

    }
    return iws, nil
}

func makePoint(info params.Parser, max bool) (ret common.Point, err error) {
    var tpl_lon, tpl_lat string

    if max {
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    } else {
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if ret.X, err = info.Float(tpl_lon, 0); err != nil {
        err = utils.WrapFmt(err, "Could not get Longitude value!")
        return
    }

    if ret.Y, err = info.Float(tpl_lat, 0); err != nil {
        err = utils.WrapFmt(err, "Could not get Latitude value!")
        return
    }

    return ret, nil
}

type(
    IWInfo struct {
        nburst int
        extent common.Rectangle
        bursts [nMaxBurst]float64
    }
    
    IWInfos [maxIW]IWInfo
)

var parCmd = common.Must("par_S1_SLC")

func iwInfo(path string) (ret IWInfo, err error) {
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    
    err = utils.NotEmpty(path, "path to annotation file")
    if err != nil {
        return
    }
    
    // Check(err, "Failed to retreive IW number from %s", path);

    par := "tmp"
    defer os.Remove(par)

    if err != nil { return }

    TOPS_par := par + ".TOPS_par"

    _, err = parCmd.Call(nil, path, nil, nil, par, nil, TOPS_par)

    if err != nil {
        err = utils.WrapFmt(err, "failed to import parameter files from '%s'",
            path)
        return
    }

    _info, err := burstCorners.Call(par, TOPS_par)

    if err != nil {
        err = utils.WrapFmt(err, "failed to parse parameter files")
        return
    }

    _TOPS, err := data.NewGammaParams(TOPS_par)
    if err != nil {
        return
    }
    TOPS := _TOPS.ToParser()
    
    nburst, err := TOPS.Int("number_of_bursts", 0)

    if err != nil {
        err = utils.WrapFmt(err, "failed to retreive number of bursts")
        return
    }

    var numbers [nMaxBurst]float64

    for ii := 1; ii < nburst+1; ii++ {
        tpl := fmt.Sprintf(burstTpl, ii)

        numbers[ii-1], err = TOPS.Float(tpl, 0)

        if err != nil {
            err = utils.WrapFmt(err, "failed to get burst number: '%s'", tpl)
            return
        }
    }
    
    info := params.FromString(_info, ":").ToParser()
    max, err := makePoint(info, true)

    if err != nil {
        err = utils.WrapFmt(err, "failed to create max latlon point")
        return
    }

    min, err := makePoint(info, false)

    if err != nil {
        err = utils.WrapFmt(err, "failed to create min latlon point")
        return
    }

    return IWInfo{
        nburst: nburst,
        extent: common.Rectangle{Min: min, Max: max},
        bursts: numbers,
    }, nil
}

func inIWs(p common.Point, IWs IWInfos) bool {
    for _, iw := range IWs {
        if p.InRect(iw.extent) {
            return true
        }
    }
    return false
}

func (iw IWInfos) contains(aoi common.AOI) bool {
    sum := 0

    for _, point := range aoi {
        if inIWs(point, iw) {
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

func (s1 Zip) Names(mode, pol string) (dat, par string) {
    path := filepath.Join(s1.Root, mode)
    
    dat = filepath.Join(path, fmt.Sprintf("%s.%s", pol, mode))
    par = dat + ".par"
    
    return
}

func (s1 Zip) SLCNames(mode, pol string, ii int) (dat, par, TOPS string) {
    slcPath := filepath.Join(s1.Root, mode)

    dat = filepath.Join(slcPath, fmt.Sprintf("iw%d_%s.%s", ii, pol, mode))
    par = dat + ".par"
    TOPS = dat + ".TOPS_par"

    return
}

func (s1 Zip) SLC(pol string) (ret SLC, err error) {
    const mode = "slc"
    tab := s1.tabName(mode, pol)

    exist, err := path.Exist(tab)
    if err != nil {
        return
    }

    if !exist {
        err = utils.WrapFmt(err, "tabfile '%s' does not exist", tab)
        return
    }

    for ii := 1; ii < 4; ii++ {
        dat, par, TOPS_par := s1.SLCNames(mode, pol, ii)
        ret.IWs[ii-1] = NewIW(dat, par, TOPS_par)
    }

    ret.Tab, ret.nIW = tab, 3

    return ret, nil
}

func (s1 Zip) MLI(mode, pol string, out *base.MLI, opt *base.MLIOpt) (err error) {
    //path := filepath.Join(s1.Root, mode)

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
    if err != nil { return }
    
    err = slc.MLI(out, opt)
    
    return
}

func (s1 Zip) tabName(mode, pol string) string {
    return filepath.Join(s1.Root, mode, fmt.Sprintf("%s.tab", pol))
}

const (
    ExtractErr utils.Werror = "failed to extract %s file from '%s'"
)

var parS1SLC = common.Must("par_S1_SLC")

func (s1 Zip) ImportSLC(dst string) (err error) {
    var _annot, _calib, _tiff, _noise string
    
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        return
    }

    defer ext.Close()

    pol := s1.pol
    tab := s1.tabName("slc", pol)

    file, err := stream.Create(tab)
    if err != nil {
        return
    }
    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        _annot = ext.Extract(annot, ii)
        _calib = ext.Extract(calib, ii)
        _tiff  = ext.Extract(tiff, ii)
        _noise = ext.Extract(noise, ii)

        if err = ext.Wrap(); err != nil {
            return
        }

        dat, par, TOPS_par := s1.SLCNames("slc", pol, ii)

        _, err = parS1SLC.Call(_tiff, _annot, _calib, _noise, par,
            dat, TOPS_par)
        if err != nil {
            err = utils.WrapFmt(err,
                "failed to import datafiles into gamma format")
            return
        }

        line := fmt.Sprintf("%s %s %s\n", dat, par, TOPS_par)

        _, err = file.WriteString(line)

        if err != nil {
            return
        }
    }

    return nil
}

func (s1 Zip) Quicklook(dst string) (s string, err error) {
    var ext = s1.newExtractor(dst)
    if err = ext.Wrap(); err != nil {
        return
    }
    defer ext.Close()

    s = ext.Extract(quicklook, 0)
    err = ext.Wrap()

    return
}

func (s1 Zip) Date() time.Time {
    return s1.date.Center()
}

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d ByDate) Less(i, j int) bool {
    return Before(d[i], d[j])
}

func Before(one, two date.Dater) bool {
    return one.Date().Before(two.Date())
}

type DateRange struct {
    start, stop, center time.Time
}

func NewDate(df date.ParseFmt, start, stop string) (d DateRange, err error) {
    _start, err := df.Parse(start)
    if err != nil {
        return
    }

    _stop, err := df.Parse(stop)
    if err != nil {
        return
    }

    // TODO: Optional check duration, is it max or min
    delta := _start.Sub(_stop) / 2.0
    d.center = _stop.Add(delta)

    d.start = _start
    d.stop = _stop

    return d, nil
}

func (d DateRange) Start() time.Time {
    return d.start
}

func (d DateRange) Center() time.Time {
    return d.center
}

func (d DateRange) Stop() time.Time {
    return d.stop
}
