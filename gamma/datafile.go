package gamma

import (
    "fmt"
    "time"
    "os"
    fp "path/filepath"
    str "strings"
    conv "strconv"
)

type (
    DataFile interface {
        Datfile() string
        Parfile() string
        Rng() (int, error)
        Azi() (int, error)
        Int(string, int) (int, error)
        Float(string, int) (float64, error)
        PlotCmd() string
        ImageFormat() (string, error)
        //Display(disArgs) error
        //Raster(Args) error
    }
    
    dataFile struct {
        Dat   string `name:"dat" default:""`
        files []string
        Params
        time.Time
    }
    
    FakeDataFile struct {
        Dat, ImgFmt string
        RngAzi
    }
    
    Subset struct {
        Begin, Nlines int
    }
)

func NewGammaParam(path string) Params {
    return Params{Par: path, Sep: ":", contents: nil}
}

func NewDataFile(dat, par, ext string) (ret dataFile, err error) {
    if len(dat) == 0 {
        err = Handle(err, "'dat' should not be an empty string")
        return
    }
    
    ret.Dat = dat
    
    if len(ext) == 0 {
        ext = "par"
    }
    
    if len(par) == 0 {
        par = fmt.Sprintf("%s.%s", dat, ext)
    }
    
    ret.Params = NewGammaParam(par)
    ret.files = []string{dat, par}
    
    if ret.Time ,err = ret.Date(); err != nil {
        err = Handle(err, "failed to retreive date from '%s'", par)
        return
    }
    
    return ret, nil
}

// TODO: implement
func FromLine(line string) (ret DataFile, err error) {
    
    return ret, nil
}

func (d *dataFile) Exist() (ret bool, err error) {
    var exist bool
    for _, file := range d.files {
        exist, err = Exist(file)

        if err != nil {
            err = Handle(err, "stat on file '%s' failed", file)
            return
        }

        if !exist {
            return false, nil
        }
    }
    return true, nil
}

func (d dataFile) Datfile() string {
    return d.Dat
}

func (d dataFile) Parfile() string {
    return d.Par
}

func (d dataFile) Rng() (int, error) {
    return d.Int("range_samples", 0)
}

func (d dataFile) Azi() (int, error) {
    return d.Int("azimuth_lines", 0)
}

func (d dataFile) ImageFormat() (string, error) {
    return d.Param("image_format")
}

func (d dataFile) Date() (ret time.Time, err error) {

    dateStr, err := d.Param("date")
    
    if err != nil {
        err = Handle(err, "failed to retreive date from '%s'", d.Par)
        return
    }
    
    split := str.Fields(dateStr)
    
    year, err := conv.Atoi(split[0])
    
    if err != nil {
        err = Handle(err, "failed retreive year from date string '%s'", dateStr)
        return
    }
    
    var month time.Month
    
    switch split[1] {
    case "01":
        month = time.January
    case "02":
        month = time.February
    case "03":
        month = time.March
    case "04":
        month = time.April
    case "05":
        month = time.May
    case "06":
        month = time.June
    case "07":
        month = time.July
    case "08":
        month = time.August
    case "09":
        month = time.September
    case "10":
        month = time.October
    case "11":
        month = time.November
    case "12":
        month = time.December
    }
    
    day, err := conv.Atoi(split[2])
        
    if err != nil {
        err = Handle(err, "failed retreive day from date string '%s'", dateStr)
        return
    }
    
    hour, err := conv.Atoi(split[3])
        
    if err != nil {
        err = Handle(err, "failed retreive hour from date string '%s'", dateStr)
        return
    }
    
    min, err := conv.Atoi(split[4])
        
    if err != nil {
        err = Handle(err, "failed retreive minute from date string '%s'", dateStr)
        return
    }
    
    sec, err := conv.ParseFloat(split[5], 64)
        
    if err != nil {
        err = Handle(err, "failed retreive seconds from string '%s'", dateStr)
        return
    }
        
    // TODO: parse nanoseconds
    
    return time.Date(year, month, day, hour, min, int(sec), 0, time.UTC), nil
}

func (d dataFile) PlotCmd() string {
    return ""
}

type (
    SLC struct {
        dataFile
    }
    
    MLI struct {
        dataFile
    }
    
    Float struct {
        dataFile
    }
)

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

var multiLook = Gamma.Must("multi_look")

type (
    // TODO: add loff, nlines
    MLIOpt struct {
        Subset
        refTab string
        Looks RngAzi
        windowFlag bool
        ScaleExp
    }
)

func (opt *MLIOpt) Parse() {
    opt.ScaleExp.Parse()
    
    if len(opt.refTab) == 0 {
        opt.refTab = "-"
    }
    
    opt.Looks.Default()
}

func (s *SLC) MakeMLI(opt MLIOpt) (ret MLI, err error) {
    opt.Parse()
    
    tmp := ""
    
    if tmp, err = TmpFileExt("mli"); err != nil {
        err = Handle(err, "failed to create tmp file")
        return
    }
    
    if ret, err = NewMLI(tmp, ""); err != nil {
        err = Handle(err, "failed to create MLI struct")
        return
    }
    
    _, err = multiLook(s.Dat, s.Par, ret.Dat, ret.Par,
                       opt.Looks.Rng, opt.Looks.Azi,
                       opt.Subset.Begin, opt.Subset.Nlines,
                       opt.ScaleExp.Scale, opt.ScaleExp.Exp)
    
    if err != nil {
        err = Handle(err, "multi_look failed")
        return
    }
    
    return ret, nil
}

type (
    SBIOpt struct {
        NormSquintDiff float64
        Looks RngAzi
        InvWeight, Keep  bool
    }
    
    SBIOut struct {
        ifg IFG
        mli MLI
    }
)


var sbiInt = Gamma.Must("SBI_INT")

func (opt *SBIOpt) Default() {
    opt.Looks.Default()
    
    if opt.NormSquintDiff == 0.0 {
        opt.NormSquintDiff = 0.5
    }
}

func (ref SLC) SplitBeamIfg(slave SLC, opt SBIOpt) (ret SBIOut, err error) {
    opt.Default()
    
    tmp := ""
    
    if tmp, err = TmpFile(); err != nil {
        err = Handle(err, "failed to create tmp file")
        return
    }
    
    if ret.ifg, err = NewIFG(tmp + ".diff", "", "", "", ""); err != nil {
        err = Handle(err, "failed to create IFG struct")
        return
    }
    
    if ret.mli, err = NewMLI(tmp + ".mli", ""); err != nil {
        err = Handle(err, "failed to create MLI struct")
        return
    }
    
    iwflg, cflg := 0, 0
    if opt.InvWeight { iwflg = 1 }
    if opt.Keep { cflg = 1 }
    
    _, err = sbiInt(ref.Dat, ref.Par, slave.Dat, slave.Par,
                    ret.ifg.Dat, ret.ifg.Par, ret.mli.Dat, ret.mli.Par, 
                    opt.NormSquintDiff, opt.Looks.Rng, opt.Looks.Azi,
                    iwflg, cflg)
    
    if err != nil {
        err = Handle(err, "SBI_INT failed")
        return
    }
    
    return ret, nil
}

type (
    SSIMode int
    
    SSIOpt struct {
        Hgt, LtFine, OutDir string
        Mode SSIMode
        Keep bool
    }
    
    SSIOut struct {
        Ifg IFG
        Unw FakeFloat
    }
)

const (
    Ifg           SSIMode = iota
    IfgUnwrapped
)

var ssiInt = Gamma.Must("SSI_INT")

func (ref SLC) SplitSpectrumIfg(slave SLC, mli MLI, opt SSIOpt) (ret SSIOut, err error) {
    mode := 1
    
    if opt.Mode == IfgUnwrapped {
        mode = 2
    }
    
    cflg := 1
    if opt.Keep { cflg = 0 }
    
    mID, sID := ref.Format(DateShort), slave.Format(DateShort)
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    _, err = ssiInt(ref.Dat, ref.Par, mli.Dat, mli.Par, opt.Hgt, opt.LtFine,
                    slave.Dat, slave.Par, mode, mID, sID, ID, opt.OutDir, cflg)
    
    if err != nil {
        err = Handle(err, "SSI_INT failed")
        return
    }
    
    // TODO: figure out the name of the output files
    
    return ret, nil
}


func (d SLC) PlotCmd() string {
    return "SLC"
}

func (s *SLC) Raster(opt RasArgs) error {
    err := opt.Parse(s)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return rasslc(opt)
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

func (d MLI) PlotCmd() string {
    return "MLI"
}

func (m *MLI) Raster(opt RasArgs) error {
    err := opt.Parse(m)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return raspwr(opt)
}

type (
    FakeMLI struct {
        FakeDataFile
    }
    
    FakeSLC struct {
        FakeDataFile
    }
    
    FakeFloat struct {
        FakeDataFile
    }
)

func (d FakeDataFile) Rng() (int, error) {
    return d.RngAzi.Rng, nil
}

func (d FakeDataFile) Azi() (int, error) {
    return d.RngAzi.Azi, nil
}

func (d FakeDataFile) ImageFormat() (string, error) {
    return d.ImgFmt, nil
}


func Display(dat DataFile, opt DisArgs) error {
    err := opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.Must("dis" + cmd)
    
    if cmd == "SLC" {
        _, err := fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines, opt.Scale,
                      opt.Exp)
        
        if err != nil {
            return Handle(err, "failed to execute display command")
        }
    }
    return nil
}

// TODO: implement proper selection of plot command
func Raster(dat DataFile, opt RasArgs, sec string) (err error) {
    err = opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.Must("ras" + cmd)
    
    switch cmd {
        case "SLC":
            err = rasslc(opt)
            
            if err != nil {
                return
            }
            
        case "MLI":
            err = raspwr(opt)
            
            if err != nil {
                return
            }
        
        default:
            err = Handle(nil, "unrecognized command type '%s'", cmd)
            return
    }
    
    if cmd == "SLC" {
        _, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
            opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
            opt.ImgFmt, opt.HeaderSize, opt.Raster)

    } else {
        if len(sec) == 0 {
            _, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.Raster, opt.ImgFmt, opt.HeaderSize)

        } else {
            _, err = fun(opt.Datfile, sec, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.Raster, opt.ImgFmt, opt.HeaderSize, opt.Raster)
        }
    }
    
    if err != nil {
        return Handle(err, "failed to create rasterfile '%s'", opt.Raster)
    }
    
    return nil
}

func Move(path *string, dir string) error {
    src := *path
    dst := fp.Join(dir, fp.Base(src))
    
    err := os.Rename(src, dst)

    if err != nil {
        return Handle(err, "failed to move file '%s' to '%s'", src, dst)
    }
    
    *path = dst
    
    return nil
}
