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
    dataFile struct {
        Dat   string
        files []string
        Params
        date
    }
    
    FakeDataFile struct {
        Dat, ImgFmt string
        RngAzi
    }
    
    FakeMLI struct {
        FakeDataFile
    }
    
    FakeSLC struct {
        FakeDataFile
    }
    
    FakeFloat struct {
        FakeDataFile
    }
    
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

    SLC struct {
        dataFile
    }

    MLI struct {
        dataFile
    }
    
    Float struct {
        dataFile
    }
    
    // TODO: add loff, nlines
    MLIOpt struct {
        refTab string
        Looks RngAzi
        windowFlag bool
        ScaleExp
    }
)

func NewGammaParam(path string) Params {
    return Params{Par: path, sep: ":", contents: nil}
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

    return ret, nil
}

func FromLine(line string) (ret DataFile, err error) {
    
    return ret, nil
}

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
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

func (d SLC) PlotCmd() string {
    return "SLC"
}

func (d MLI) PlotCmd() string {
    return "MLI"
}

func (d FakeDataFile) Rng() (int, error) {
    return d.RngAzi.Rng, nil
}

func (d FakeDataFile) Azi() (int, error) {
    return d.RngAzi.Azi, nil
}

func (d FakeDataFile) ImageFormat() (string, error) {
    return d.ImgFmt, nil
}

func (opt *MLIOpt) Parse() {
    opt.ScaleExp.Parse()
    
    if len(opt.refTab) == 0 {
        opt.refTab = "-"
    }
    
    if opt.Looks.Rng == 0 {
        opt.Looks.Rng = 1
    }
    
    if opt.Looks.Azi == 0 {
        opt.Looks.Azi = 1
    }
}

func (s *SLC) Raster(opt rasArgs) error {
    err := opt.Parse(s)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return rasslc(opt)
}

func (m *MLI) Raster(opt rasArgs) error {
    err := opt.Parse(m)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return raspwr(opt)
}

func Display(dat DataFile, opt disArgs) error {
    err := opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("dis" + cmd)
    
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
func Raster(dat DataFile, opt rasArgs, sec string) (err error) {
    err = opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("ras" + cmd)
    
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
            opt.ImgFmt, opt.headerSize, opt.raster)

    } else {
        if len(sec) == 0 {
            _, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.raster, opt.ImgFmt, opt.headerSize)

        } else {
            _, err = fun(opt.Datfile, sec, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.raster, opt.ImgFmt, opt.headerSize, opt.raster)
        }
    }
    
    if err != nil {
        return Handle(err, "failed to create rasterfile '%s'", opt.raster)
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
