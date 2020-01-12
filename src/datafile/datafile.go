package datafile

// TODO: seperate field for storing rng, azi, DType values

import (
    "fmt"
    "time"
    "os"
    "path/filepath"
    "strings"
    "strconv"
    "time"
    
    "../utils"
    "../common"
)

const merr = utils.ModuleName("gamma.datafile")

type (    
    IDatFile interface {
        Datfile() string
        Rng() int
        Azi() int
        Dtype() DType
        //Move(string) (DatFile, error)
    }

    DatFile struct {
        Dat     string
        Ra      common.RngAzi
        Time time.Time
        DType
        Params
    }
)


func (d DatFile) Rng() int {
    return d.Ra.Rng
}

func (d DatFile) Azi() int {
    return d.Ra.Azi
}

func (d DatFile) TypeCheck(ftype, expect string, dtypes... DType) (err error) {
    b, D := false, d.DType
    
    for _, dt := range dtypes {
        if D == dt {
            b = true
            break
        }
    }
    
    if !b {
        err = TypeMismatchError{ftype:ftype, expected:expect, DType:D}
        return
    }
    
    return nil
}

func New(rng, azi int, dtype DType) (d DatFile, err error) {
    ferr := merr.Make("New")
    
    return
}

func FromFile(path, rng, azi, dtype string) (d DatFile, err error) {
    ferr := merr.Make("FromFile")
    
    d.Params, err = NewGammaParam(path)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    d.Dat, err = d.Param("datafile")
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    d.Ra.Rng, err = d.ParseRng(rng)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }

    d.Ra.Azi, err = d.ParseAzi(azi)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    d.DType, err = d.ParseDtype(dtype)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    return
}

func (d DatFile) Save() (err error) {
    ferr := merr.Make("DatFile.Save")
    
    file, err := os.Create(d.Params.filePath)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer file.Close()
    
    d.params["datafile"] = d.Dat
    
    err = d.Params.Save(file)
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (d DatFile) SaveDat() (err error) {
    ferr := merr.Make("DatFile.Save")
    
    file, err := os.Open(d.Params.filePath)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer file.Close()
    
    s := fmt.Sprintf("datafile: %s", d.Dat)
    
    _, err = file.WriteString(s)
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (d DatFile) ParseRng(key string) (i int, err error) {
    i, err = d.Int(key, 0)
    if err != nil {
        err = merr.Make("DatFile.ParseRng").Wrap(err)
    }
    return
}

func (d DatFile) ParseAzi(key string) (i int, err error) {
    i, err = d.Int(key, 0)
    if err != nil {
        err = merr.Make("DatFile.ParseAzi").Wrap(err)
    }
    return
}

func (df DatFile) ParseDtype(key string) (d DType, err error) {
    var ferr = merr.Make("DatParFile.ParseDtype")
    
    s, err := df.Param(key)
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    d.Set(s)
    
    return d, nil
}

func (d DatFile) Like(name string, dtype DType) (ret DatFile, err error) {
    ferr := merr.Make("DatFile.Like")
    
    if dtype == Unknown {
        dtype = d.DType
    }
    
    if name, err = filepath.Abs(name); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if ret, err = NewDatFile(name, dtype); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    ret.Ra = d.Ra
    return ret, nil
}

func (d DatFile) Datfile() string {
    return d.Dat
}

func (d DatFile) Dtype() DType {
    return d.DType
}


func (d DatFile) Move(dir string) (dm DatFile, err error) {
    var ferr = merr.Make("DatFile.Move")
    
    if dm.Dat, err = Move(d.Dat, dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    dm.Ra, dm.DType = d.Ra, d.DType
    
    return dm, nil
}

func (d DatFile) Exist() (b bool, err error) {
    var ferr = merr.Make("DatFile.Exist")
    
    if b, err = utils.Exist(d.Dat); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    return b, nil
}

func SameCols(one IDatFile, two IDatFile) (err error) {
    var ferr = merr.Make("SameCols")
    
    n1, n2 := one.Rng(), two.Rng()
    
    if n1 != n2 {
        return ferr.Wrap(ShapeMismatchError{
            dat1: one.Datfile(),
            dat2: two.Datfile(),
            n1:n1,
            n2:n2,
            dim: "range samples / columns",
        })
    }
    return nil
}

func SameRows(one IDatFile, two IDatFile) error {
    var ferr = merr.Make("SameRows")
    
    n1, n2 := one.Azi(), two.Azi()
    
    if n1 != n2 {
        return ferr.Wrap(ShapeMismatchError{
            dat1: one.Datfile(),
            dat2: two.Datfile(),
            n1:n1,
            n2:n2,
            dim: "azimuth lines / rows",
        })
    }
    
    return nil
}

func SameShape(one IDatFile, two IDatFile) (err error) {
    var ferr = merr.Make("SameShape")
    
    if err = SameCols(one, two); err != nil {
        return ferr.Wrap(err)
    }

    if err = SameRows(one, two); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

//func (d *DatFile) Parse() (err error) {
    //var ferr = merr.Make("DatParFile.Parse")
    
    //if d.Ra.Rng, err = d.ParseRng(); err != nil {
        //return ferr.Wrap(err)
    //}
    
    //if d.Ra.Azi, err = d.ParseAzi(); err != nil {
        //return ferr.Wrap(err)
    //}
    
    //if d.Time, err = d.ParseDate(); err != nil {
        //return ferr.Wrap(err)
    //}
    
    //return nil
//}

func (d DatFile) ParseDate() (t time.Time, err error) {
    var ferr = merr.Make("DatParFile.ParseDate")
    
    dateStr, err := d.Param("date")
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    split := strings.Fields(dateStr)
    
    year, err := strconv.Atoi(split[0])
    
    if err != nil {
        err = ferr.Wrap(TimeParseErr.Wrap(err, "year", dateStr))
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
        err = ferr.Wrap(TimeParseErr.Wrap(err, "day", dateStr))
        return
    }
    
    var (
        hour, min int
        sec float64
    )
    
    if len(split) == 6 {
        
        hour, err = strconv.Atoi(split[3])
            
        if err != nil {
            err = ferr.Wrap(TimeParseErr.Wrap(err, "hour", dateStr))
            return
        }
        
        min, err = strconv.Atoi(split[4])
            
        if err != nil {
            err = ferr.Wrap(TimeParseErr.Wrap(err, "minute", dateStr))
            return
        }
        
        sec, err = strconv.ParseFloat(split[5], 64)
            
        if err != nil {
            err = ferr.Wrap(TimeParseErr.Wrap(err, "seconds", dateStr))
            return
        }
    }        
    // TODO: parse nanoseconds
    
    t = time.Date(year, month, day, hour, min, int(sec), 0, time.UTC)
    
    return t, nil
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

func Move(path string, dir string) (s string, err error) {
    var ferr = merr.Make("Move")
    
    dst, err := filepath.Abs(filepath.Join(dir, filepath.Base(path)))
    if err != nil {
        err = ferr.WrapFmt(err, "failed to create absolute path")
        return
    }
    
    if err = os.Rename(path, dst); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    return dst, nil
}


func (d *DatFile) Set(s string) (err error) {
    return LoadJson(s, d)
}
