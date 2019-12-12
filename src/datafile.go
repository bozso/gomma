package gamma

import (
    "fmt"
    "time"
    "os"
    "encoding/json"
    "path/filepath"
    "strings"
    "strconv"
    //"log"
    //"reflect"
)

type (    
    Serialize interface {
        jsonMap() JSONMap
        jsonName() string
        FromJson(JSONMap) error        
    }
    
    IDatFile interface {
        Datfile() string
        Rng() int
        Azi() int
        Dtype() DType
        //Move(string) (DatFile, error)
        Raster(opt RasArgs) error
    }
    
    IDatParFile interface {
        IDatFile
        ParseRng() (int, error)
        ParseAzi() (int, error)
        ParseDate() (time.Time, error)
        ParseFmt() (string, error)
        TimeStr(dateFormat) string
        
        // TODO: implement this
        // ParseDType() (DType, error)
    }
    
    DType int
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
    Any
)

func (d *DType) Decode(s string) error {
    in := strings.ToUpper(s)
    
    switch in {
    case "FLOAT":
        *d = Float
    case "DOUBLE":
        *d = Double
    case "SCOMPLEX":
        *d = ShortCpx
    case "FCOMPLEX":
        *d = FloatCpx
    case "SUN", "RASTER", "BMP":
        *d = Raster
    case "UNSIGNED CHAR":
        *d = UChar
    case "SHORT":
        *d = Short
    case "ANY":
        *d = Any
    default:
        *d = Unknown
    }
    
    return nil
}

func (d DType) String() string {
    switch d {
    case Float:
        return "FLOAT"
    case Double:
        return "DOUBLE"
    case ShortCpx:
        return "SCOMPLEX"
    case FloatCpx:
        return "FCOMPLEX"
    case Raster:
        return "RASTER"
    case UChar:
        return "UNSIGNED CHAR"
    case Short:
        return "SHORT"
    case Any:
        return "ANY"
    default:
        return "UNKNOWN"
    }
}

type ZeroDimError struct {
    dim string
    Err error
}


func (e ZeroDimError) Error() string {
    return fmt.Sprintf("expected %s to be non zero", e.dim)
}

func (e ZeroDimError) Unwrap() error {
    return e.Err
}

type RngAzi struct {
    Rng int `json:"rng" name:"rng" default:"0"`
    Azi int `json:"azi" name:"azi" default:"0"`
}

func (ra *RngAzi) Decode(s string) (err error) {
    if len(s) == 0 {
        return EmptyStringError
    }
    
    split := NewSplitParser(s, ",")
    
    ra.Rng = split.Int(0)
    ra.Azi = split.Int(1)
    
    return split.err
}

func (ra RngAzi) Check() (err error) {
    if ra.Rng == 0 {
        return ZeroDimError{dim: "range samples / columns"}
    }
    
    if ra.Azi == 0 {
        return ZeroDimError{dim: "azimuth lines / rows"}
    }
    return nil
}


type TypeMismatchError struct {
    ftype, expected string
    DType
    Err error
}

func (e TypeMismatchError) Error() string {
    return fmt.Sprintf("expected datatype '%s' for %s datafile, got '%s'",
        e.expected, e.ftype, e.DType.String())
}

func (e TypeMismatchError) Unwrap() error {
    return e.Err
}

type UnknownTypeError struct {
    DType
    Err error
}

func (e UnknownTypeError) Error() string {
    return fmt.Sprintf("unrecognised type '%s', expected a valid datatype",
        e.DType.String())
}

func (e UnknownTypeError) Unwrap() error {
    return e.Err
}

type WrongTypeError struct {
    DType
    kind string
    Err error
}

func (e WrongTypeError) Error() string {
    return fmt.Sprintf("wrong datatype '%s' for %s", e.kind, e.DType.String())
}

func (e WrongTypeError) Unwrap() error {
    return e.Err
}


func NewGammaParam(path string) Params {
    return Params{Par: path, Sep: ":", contents: nil}
}

const (
    StructCreateError Werror = "failed to create %s Struct"
)

type (
    URngAzi struct {
        rng int `name:"rng" default:"0"`
        azi int `name:"azi" default:"0"`
    }
    
    DatFile struct {
        Dat string `name:"dat"`
        URngAzi
        DType
    }
)

func NewDatFile(path string, dt DType) (ret DatFile, err error) {
    if len(path) == 0 {
        err = fmt.Errorf("expected datafile path to be a non empty string")
        return
    }
    
    return DatFile{Dat: path, DType: dt}, nil
}

func TmpDatFile(ext string, dt DType) (ret DatFile, err error) {
    var dat string
    if dat, err = TmpFile(ext); err != nil {
        return
    }
    
    ret.Dat = dat
    ret.DType = dt
    
    return ret, nil
}

func (d DatFile) Like(name string, dtype DType) (ret DatFile, err error) {
    if dtype == Unknown {
        dtype = d.DType
    }
    
    if name, err = filepath.Abs(name); err != nil {
        return
    }
    
    if ret, err = NewDatFile(name, dtype); err != nil {
        return
    }
    
    ret.URngAzi = d.URngAzi
    return ret, nil
}

func (d DatFile) Datfile() string {
    return d.Dat
}

func (d DatFile) Rng() int {
    return d.rng
}

func (d DatFile) Azi() int {
    return d.azi
}

func (d DatFile) Dtype() DType {
    return d.DType
}

func (d DatFile) jsonName() string {
    return d.Dat + ".json"
}

func (d DatFile) jsonMap() JSONMap {
    return JSONMap{
        "datafile": d.Dat,
        "range_samples": d.rng,
        "azimuth_lines": d.azi,
        "dtype": d.DType.String(),
    }
}

func (d *DatFile) FromJson(m JSONMap) (err error) {
    if d.Dat, err = m.String("datafile"); err != nil {
        err = Handle(err, "failed to retreive datafile")
        return
    }
    
    var dt string
    if dt, err = m.String("dtype"); err != nil {
        err = Handle(err, "failed to retreive dtype")
        return
    }
    
    d.Decode(dt)
    
    if d.rng, err = m.Int("range_samples"); err != nil {
        //err = RngError.Wrap(err, path)
        return
    }
    
    if d.azi, err = m.Int("azimuth_lines"); err != nil {
        //err = AziError.Wrap(err, path)
        return
    }
    
    return nil
}

func (d DatFile) Move(dir string) (ret DatFile, err error) {
    if ret.Dat, err = Move(d.Dat, dir); err != nil {
        return
    }
    
    ret.URngAzi, ret.DType = d.URngAzi, d.DType
    
    return ret, nil
}

func (d DatFile) Exist() (ret bool, err error) {
    ret, err = Exist(d.Dat)
    return
}

type ShapeMismatchError struct {
    dat1, dat2, dim string
    n1, n2 int
}

func (s ShapeMismatchError) Error() string {
    return fmt.Sprintf("expected datafile '%s' to have the same %s as " + 
                       "datafile '%s' (%d != %d)", s.dat1, s.dim, s.dat2, s.n1,
                       s.n2)
}

func SameCols(one IDatFile, two IDatFile) error {
    n1, n2 := one.Rng(), two.Rng()
    
    if n1 != n2 {
        return ShapeMismatchError{
            dat1: one.Datfile(),
            dat2: two.Datfile(),
            n1:n1,
            n2:n2,
            dim: "range samples / columns",
        }
    }
    return nil
}

func SameRows(one IDatFile, two IDatFile) error {
    n1, n2 := one.Azi(), two.Azi()
    
    if n1 != n2 {
        return ShapeMismatchError{
            dat1: one.Datfile(),
            dat2: two.Datfile(),
            n1:n1,
            n2:n2,
            dim: "azimuth lines / rows",
        }
    }
    return nil
}

func SameShape(one IDatFile, two IDatFile) (err error) {
    if err = SameCols(one, two); err != nil {
        return
    }
    
    return SameRows(one, two)
}
    
type DatParFile struct {
    DatFile
    Params
    time.Time `json:"-"`
}

func NewDatParFile(dat, par, ext string, dt DType) (ret DatParFile, err error) {
    if ret.DatFile, err = NewDatFile(dat, dt); err != nil {
        return
    }
    
    if len(par) == 0 {
        par = fmt.Sprintf("%s.%s", dat, ext)
    }
    
    ret.Par = par
    ret.Sep = ":"
    
    return ret, nil
}

func TmpDatParFile(ext string, parExt string, dt DType) (ret DatParFile, err error) {
    if ret.DatFile, err = TmpDatFile(ext, dt); err != nil {
        return
    }
    
    if len(parExt) == 0 {
        parExt = "par"
    }
    
    ret.Par = fmt.Sprintf("%s.%s", ret.Dat, parExt)
    ret.Sep = ":"
    
    return ret, nil
}

func (d *DatParFile) Parse() (err error) {
    if d.rng, err = d.ParseRng(); err != nil {
        return
    }
    
    if d.azi, err = d.ParseAzi(); err != nil {
        return
    }
    
    if d.Time, err = d.ParseDate(); err != nil {
        return
    }
    return nil
}

func (d DatParFile) Move(dir string) (ret DatParFile, err error) {
    if ret.DatFile, err = d.DatFile.Move(dir); err != nil {
        return
    }
    
    if ret.Par, err = Move(d.Par, dir); err != nil {
        return
    }
    
    return ret, nil
}

func (d DatParFile) Exist() (ret bool, err error) {
    var de, pe bool
    
    if de, err = d.DatFile.Exist(); err != nil {
        return
    }
    
    if pe, err = Exist(d.Par); err != nil {
        return
    }
    
    return de && pe, nil
}

func (d DatParFile) jsonMap() JSONMap {
    ret := d.DatFile.jsonMap()
    ret["parameterfile"] = d.Par
    
    return ret
}

func (d *DatParFile) FromJson(m JSONMap) (err error) {
    if err = d.DatFile.FromJson(m); err != nil {
        return
    }
    
    if d.Par, err = m.String("parameterfile"); err != nil {
        err = Handle(err, "failed to retreive paramfile")
        return
    }
    d.Sep = ":"
    
    return nil    
}

const (
    RngError Werror = "failed to retreive range samples from '%s'"
    AziError Werror = "failed to retreive azimuth lines from '%s'"
)

func (d DatParFile) TimeStr(format dateFormat) string {
    switch format {
    case DShort:
        return d.Time.Format(DateShort)
    case DLong:
        return d.Time.Format(DateLong)
    }
    return ""
}

func (d DatParFile) ParseRng() (int, error) {
    return d.Int("range_samples", 0)
}

func (d DatParFile) ParseAzi() (int, error) {
    return d.Int("azimuth_lines", 0)
}

func (d DatParFile) ParseFmt() (string, error) {
    return d.Param("image_format")
}

func (d DatParFile) ParseDtype() (dt DType, err error) {
    var s string
    
    if s, err = d.Param("image_format"); err != nil {
        return
    }
    
    dt.Decode(s)
    
    if dt == Unknown {
        err = fmt.Errorf("failed to determine data type based on '%s'",
            s)
        return
    }
    
    return dt, nil
}

const (
    TimeParseErr Werror = "failed retreive %s from date string '%s'"
)

func (d DatParFile) ParseDate() (ret time.Time, err error) {

    dateStr, err := d.Param("date")
    
    if err != nil {
        err = Handle(err, "failed to retreive date from '%s'", d.Par)
        return
    }
    
    split := strings.Fields(dateStr)
    
    year, err := strconv.Atoi(split[0])
    
    if err != nil {
        err = TimeParseErr.Wrap(err, "year", dateStr)
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
    
    
    day, err := strconv.Atoi(split[2])
        
    if err != nil {
        err = TimeParseErr.Wrap(err, "day", dateStr)
        return
    }
    
    var (
        hour, min int
        sec float64
    )
    
    if len(split) == 6 {
        
        hour, err = strconv.Atoi(split[3])
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "hour", dateStr)
            return
        }
        
        min, err = strconv.Atoi(split[4])
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "minute", dateStr)
            return
        }
        
        sec, err = strconv.ParseFloat(split[5], 64)
            
        if err != nil {
            err = TimeParseErr.Wrap(err, "seconds", dateStr)
            return
        }
    }        
    // TODO: parse nanoseconds
    
    return time.Date(year, month, day, hour, min, int(sec), 0, time.UTC), nil
}

func ID(one IDatParFile, two IDatParFile, format dateFormat) string {
    return fmt.Sprintf("%s_%s", one.TimeStr(format), two.TimeStr(format))
}

//func Display(dat DataFile, opt DisArgs) error {
    //err := opt.Parse(dat)
    
    //if err != nil {
        //return Handle(err, "failed to parse display options")
    //}
    
    //cmd := opt.Cmd
    //fun := Gamma.Must("dis" + cmd)
    
    //if cmd == "SLC" {
        //_, err := fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines, opt.Scale,
                      //opt.Exp)
        
        //if err != nil {
            //return Handle(err, "failed to execute display command")
        //}
    //}
    //return nil
//}

// TODO: implement proper selection of plot command
//func (d DatFile) Raster(opt RasArgs) (err error) {
    //err = opt.Parse(d)
    
    //if err != nil {
        //return Handle(err, "failed to parse display options")
    //}
    //
    //cmd := opt.Cmd
    //fun := Gamma.Must("ras" + cmd)
    
    //switch cmd {
        //case "SLC":
            //err = rasslc(opt)
            
            //if err != nil {
                //return
            //}
            
        //case "MLI":
            //err = raspwr(opt)
            
            //if err != nil {
                //return
            //}
        
        //default:
            //err = Handle(nil, "unrecognized command type '%s'", cmd)
            //return
    //}
    
    //if cmd == "SLC" {
        //_, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
            //opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
            //opt.ImgFmt, opt.HeaderSize, opt.Raster)

    //} else {
        //if len(sec) == 0 {
            //_, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                //opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                //opt.LR, opt.Raster, opt.ImgFmt, opt.HeaderSize)

        //} else {
            //_, err = fun(opt.Datfile, sec, opt.Rng, opt.Start, opt.Nlines,
                //opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                //opt.LR, opt.Raster, opt.ImgFmt, opt.HeaderSize, opt.Raster)
        //}
    //}
    
    //if err != nil {
        //return Handle(err, "failed to create rasterfile '%s'", opt.Raster)
    //}
    //
    //return nil
//}

type(
    Subset struct {
        RngOffset int `name:"roff" default:"0"`
        AziOffset int `name:"aoff" default:"0"`
        RngWidth int `name:"rwidth" default:"0"`
        AziLines int `name:"alines" default:"0"`
    }
)

//func (s *Subset) Parse(d IDatParFile) {
    //if s.RngWidth == 0 {
        //s.RngWidth = d.GetRng()
    //}
    
    //if s.AziLines == 0 {
        //s.AziLines = d.GetAzi()
    //}
//}

func Move(path string, dir string) (ret string, err error) {
    dst, err := filepath.Abs(filepath.Join(dir, filepath.Base(path)));
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

func Save(path string, d Serialize) (err error) {
    if len(path) == 0 {
        if path, err = filepath.Abs(d.jsonName()); err != nil {
            return
        }
    }
    
    return SaveJson(path, d.jsonMap())
}

func Load(path string, d Serialize) (err error) {
    
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
    
    return d.FromJson(m) 
}
