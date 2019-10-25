package gamma

import (
    "fmt"
    "time"
    "os"
    //"log"
    "encoding/json"
    fp "path/filepath"
    str "strings"
    conv "strconv"
    //ref "reflect"
)

type (
    DataFile interface {
        Datfile() string
        Parfile() string
        GetRng() int
        GetAzi() int
        GetDate() time.Time
        TimeStr(dateFormat) string
        TypeStr() string
        Int(string, int) (int, error)
        Float(string, int) (float64, error)
        PlotCmd() string
        //ImageFormat() (string, error)
        Move(string) (DataFile, error)
        //Display(disArgs) error
        Raster(RasArgs) error
        Save(string) error
    }
    
    DType int
    
    dataFile struct {
        Dat   string    `json:"datafile" name:"dat"`
        Dtype DType
        Params
        RngAzi
        time.Time       `json:"-"`
    }
)


const (
    Float DType = iota
    Double
    ShortCpx
    FloatCpx
    Raster
    UChar
    Short
    Unknown
)

func str2dtype(in string) (ret DType, err error) {
    in = str.ToUpper(in)
    
    switch in {
    case "FLOAT":
        return Float, nil
    case "DOUBLE":
        return Double, nil
    case "SCOMPLEX":
        return ShortCpx, nil
    case "FCOMPLEX":
        return FloatCpx, nil
    case "SUN", "RASTER", "BMP":
        return Raster, nil
    case "UNSIGNED CHAR":
        return UChar, nil
    case "SHORT":
        return Short, nil
    default:
        return ret, fmt.Errorf("unrecognized data type: '%s'", in)
    }
}

func dtype2str(in DType) (ret string, err error) {
    switch in {
    case Float:
        return "FLOAT", nil
    case Double:
        return "DOUBLE", nil
    case ShortCpx:
        return "SCOMPLEX", nil
    case FloatCpx:
        return "FCOMPLEX", nil
    case Raster:
        return "RASTER", nil
    case UChar:
        return "UNSIGNED CHAR", nil
    case Short:
        return "SHORT", nil
    default:
        return ret, fmt.Errorf("unrecognized data type: '%s'", in)
    }
}


func NewGammaParam(path string) Params {
    return Params{Par: path, Sep: ":", contents: nil}
}

const (
    DataCreateErr Werror = "failed to create %s Struct"
)

func Newdatafile(dat, par string) (ret dataFile, err error) {
    if len(dat) == 0 {
        err = Handle(err, "'dat' should not be an empty string")
        return
    }
    
    ret.Dat = dat
    
    if len(par) == 0 {
        par = dat + ".par"
    }
    
    ret.Params = NewGammaParam(par)
    
    return ret, nil
}

func TmpDataFile() (ret dataFile, err error) {
    dat, err := TmpFileExt("dat")
    if err != nil {
        return ret, err
    }
    
    return Newdatafile(dat, "")
}

const (
    RngError Werror = "failed to retreive range samples from '%s'"
    AziError Werror = "failed to retreive azimuth lines from '%s'"
)


func NewDataFile(dat, par string, dt DType) (ret dataFile, err error) {
    if ret, err = Newdatafile(dat, par); err != nil {
        err = DataCreateErr.Wrap(err, "dataFile")
        return
    }
    
    if ret.Rng, err = ret.rng(); err != nil {
        err = RngError.Wrap(err, par)
        return
    }
    
    if ret.Azi, err = ret.azi(); err != nil {
        err = AziError.Wrap(err, par)
        return
    }
    
    if ret.Time, err = ret.Date(); err != nil {
        err = Handle(err, "failed to retreive date from '%s'", par)
        return
    }
    
    if dt == Unknown {
        var format string
        if format, err = ret.imgfmt(); err != nil {
            err = Handle(err, "failed to retreive image format from '%s'", par)
            return
        }
        
        if ret.Dtype, err = str2dtype(format); err != nil {
            err = Handle(err, "failed to retreive datatype")
            return
        }
    } else {
        ret.Dtype = dt
    }
    
    return ret, nil
}

// TODO: implement
func FromLine(line string) (ret DataFile, err error) {
    
    return ret, nil
}

func (d dataFile) Exist() (ret bool, err error) {
    if ret, err = Exist(d.Dat); err != nil {
        //err = Handle(err, "stat on file '%s' failed", d.Dat)
        return
    }

    if !ret {
        return false, nil
    }
    
    if ret, err = Exist(d.Par); err != nil {
        //err = Handle(err, "stat on file '%s' failed", d.Par)
        return
    }

    return ret, nil
}

func (d dataFile) Datfile() string {
    return d.Dat
}

func (d dataFile) Parfile() string {
    return d.Par
}

func (d dataFile) GetRng() int {
    return d.Rng
}

func (d dataFile) GetAzi() int {
    return d.Azi
}

func (d dataFile) GetDate() time.Time {
    return d.Time
}

func (d dataFile) TimeStr(format dateFormat) string {
    switch format {
    case DShort:
        return d.Time.Format(DateShort)
    case DLong:
        return d.Time.Format(DateLong)
    }
    return ""
}

func (d dataFile) rng() (int, error) {
    return d.Int("range_samples", 0)
}

func (d dataFile) azi() (int, error) {
    return d.Int("azimuth_lines", 0)
}

func (d dataFile) imgfmt() (string, error) {
    return d.Param("image_format")
}

func ID(one DataFile, two DataFile, format dateFormat) string {
    return fmt.Sprintf("%s_%s", one.TimeStr(format), two.TimeStr(format))
}

const (
    TimeParseErr Werror = "failed retreive %s from date string '%s'"
)

func (d dataFile) Date() (ret time.Time, err error) {

    dateStr, err := d.Param("date")
    
    if err != nil {
        err = Handle(err, "failed to retreive date from '%s'", d.Par)
        return
    }
    
    split := str.Fields(dateStr)
    
    year, err := conv.Atoi(split[0])
    
    if err != nil {
        err = TimeParseErr.Wrap(err, "year", dateStr)
        //err = Handle(err, "failed retreive year from date string '%s'", dateStr)
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
        err = TimeParseErr.Wrap(err, "day", dateStr)
        return
    }
    
    var (
        hour, min int
        sec float64
    )
    
    if len(split) == 6 {
        
        hour, err = conv.Atoi(split[3])
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "hour", dateStr)
            return
        }
        
        min, err = conv.Atoi(split[4])
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "minute", dateStr)
            return
        }
        
        sec, err = conv.ParseFloat(split[5], 64)
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "seconds", dateStr)
            return
        }
    }        
    // TODO: parse nanoseconds
    
    return time.Date(year, month, day, hour, min, int(sec), 0, time.UTC), nil
}

func (d dataFile) TypeStr() string {
    return "Unknown"
}

func (d dataFile) PlotCmd() string {
    return ""
}

func (d dataFile) Save(path string) error {
    return SaveJson(path, &d)
}

func (d dataFile) Move(dir string) (ret DataFile, err error) {
    var dat, par string
    
    if dat, err = Move(d.Dat, dir); err != nil {
        return
    }
    
    if par, err = Move(d.Par, dir); err != nil {
        return
    }
    
    if ret, err = NewDataFile(dat, par, d.Dtype); err != nil {
        err = DataCreateErr.Wrap(err, "DataFile")
        return
    }
    
    return ret, nil
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
func (d dataFile) Raster(opt RasArgs) (err error) {
    err = opt.Parse(d)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    /*
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
    */
    return nil
}

type(
    Subset struct {
        RngOffset int `name:"roff" default:"0"`
        AziOffset int `name:"aoff" default:"0"`
        RngWidth int `name:"rwidth" default:"0"`
        AziLines int `name:"alines" default:"0"`
    }
)

func (s *Subset) Parse(d DataFile) {
    if s.RngWidth == 0 {
        s.RngWidth = d.GetRng()
    }
    
    if s.AziLines == 0 {
        s.AziLines = d.GetAzi()
    }
}

func Move(path string, dir string) (ret string, err error) {
    dst, err := fp.Abs(fp.Join(dir, fp.Base(path)));
    if err != nil {
        err = Handle(err, "failed to create absolute path")
        return
    }
    
    if err = os.Rename(path, dst); err != nil {
        err = MoveErr.Wrap(err, path, dst)
        //err = Handle(err, "failed to move file '%s' to '%s'", path, dst)
        return
    }
    
    return dst, nil
}

// TODO: implement dtype2str
func (d dataFile) MarshalJSON() ([]byte, error) {
    dt, err := dtype2str(d.Dtype)
    if err != nil {
        return nil, err
    }
    
    return json.Marshal(JSONMap{
        "datafile": d.Dat,
        "paramfile": d.Par,
        "range_samples": d.Rng,
        "azimuth_lines": d.Azi,
        "dtype": dt,
        "ftype": d.TypeStr(),
    })
}

func LoadDataFile(path string) (ret DataFile, err error) {
    
    data, err := ReadFile(path)
    if err != nil {
        err = Handle(err, "failed to read file '%s'", path)
        return
    }
    
    m := make(JSONMap)
    
    if err = json.Unmarshal(data, &m); err != nil {
        err = Handle(err, "failed to parse json data %s'", data)
        return
    }
    
    ti, ok := m["ftype"]
    if !ok {
        err = KeyErr.Make("ftype", m)
        return
    }
    
    t, ok := ti.(string)
    if !ok {
        err = TypeErr.Make(t, ti, "string")
        return
    }
    
    
    var (
        dat, par, dt, quality, diffPar, simUnwrap string
        ra RngAzi
    )
    
    t = str.ToUpper(t)
    
    switch t {
    case "SLC", "MLI":
        if dat, err = m.String("datafile"); err != nil {
            err = Handle(err, "failed to retreive datafile")
            return
        }
        
        if par, err = m.String("paramfile"); err != nil {
            err = Handle(err, "failed to retreive paramfile")
            return
        }
        
        if dt, err = m.String("dtype"); err != nil {
            err = Handle(err, "failed to retreive dtype")
            return
        }
        
        if ra.Rng, err = m.Int("range_samples"); err != nil {
            err = RngError.Wrap(err, path)
            return
        }
        
        if ra.Azi, err = m.Int("azimuth_lines"); err != nil {
            err = AziError.Wrap(err, path)
            return
        }
        
        switch t {
        case "IFG":
            if quality, err = m.String("quality"); err != nil {
                err = Handle(err, "failed to retreive quality file")
                return
            }
            
            if diffPar, err = m.String("diffparfile"); err != nil {
                err = Handle(err, "failed to diffparfile")
                return
            }
            
            if simUnwrap, err = m.String("simulated_unwrapped"); err != nil {
                err = Handle(err, "failed to simulated unwrapped datafile")
                return
            }
        }
    }
        
    switch t {
    case "SLC":
        if ret, err = NewSLC(dat, par); err != nil {
            err = DataCreateErr.Wrap(err, "SLC")
            //err = Handle(err, "failed to create SLC struct")
            return
        }
    case "MLI":
        if ret, err = NewMLI(dat, par); err != nil {
            err = DataCreateErr.Wrap(err, "MLI")
            //err = Handle(err, "failed to create MLI struct")
            return
        }
    case "IFG":
        ret, err = NewIFG(dat, par, simUnwrap, diffPar, quality)
        if err != nil {
            err = DataCreateErr.Wrap(err, "IFG")
            //err = Handle(err, "failed to create IFG struct")
            return
        }
    default:
        var dtype DType
        if dtype, err = str2dtype(dt); err != nil {
            err = Handle(err, "failed to retreive datatype from string '%s'",
                dt)
            return
        }
        
        if ret, err = NewDataFile(dat, par, dtype); err != nil {
            err = DataCreateErr.Wrap(err, "DataFile")
            return
        }
    }
    
    return ret, nil
}
