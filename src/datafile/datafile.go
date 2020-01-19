package datafile

// TODO: seperate field for storing rng, azi, DType values

import (
    //"fmt"
    "os"
    "path/filepath"
    "strings"
    "strconv"
    "time"
    
    "../utils"
    "../common"
)

const merr = utils.ModuleName("gamma.datafile")

const (
    keyDatafile = "golang_meta_datafile"
    keyRng = "golang_meta_rng"
    keyAzi = "golang_meta_azi"
    keyDtype = "golang_meta_dtype"
    separator = ":"
)

type (    
    IFile interface {
        FilePath() string
        Rng() int
        Azi() int
        Dtype() DType
        //Move(string) (File, error)
    }
    
    File struct {
        Dat, Par string
        ra       common.RngAzi
        Time     time.Time
        DType
    }
)

func New(dat string, rng, azi int, dtype DType) (d File) {
    d.Dat, d.Par = dat, dat + ".par"
    d.ra.Rng, d.ra.Azi, d.DType = rng, azi, dtype
    
    return
}

func FromFile(path string) (d File, err error) {
    d.Par = path
    
    pr := NewReader(path, separator)
    
    d.Dat = pr.Param(keyDatafile)
    
    d.ra.Rng = pr.Int(keyRng, 0)
    d.ra.Azi = pr.Int(keyAzi, 0)
    
    ds := pr.Param(keyDtype)
    
    if err = pr.Wrap(); err != nil {
        return
    }
    
    err = d.DType.Set(ds)
    
    return
}

func (d File) Rng() int {
    return d.ra.Rng
}

func (d File) Azi() int {
    return d.ra.Azi
}

func (d File) FilePath() string {
    return d.Dat
}

func (d File) Dtype() DType {
    return d.DType
}

func (d File) TypeCheck(ftype, expect string, dtypes... DType) (err error) {
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

func (d File) Save() (err error) {
    path := d.Par
    
    exists, err := utils.Exist(path)
    if err != nil {
        return
    }
    
    var p params
    if !exists {
        p = make(params)
    } else {
        p, err = fromFile(path, separator)
        if err != nil {
            return
        }
    }
    
    p[keyDatafile] = d.Dat
    p[keyRng] = strconv.Itoa(d.ra.Rng)
    p[keyAzi] = strconv.Itoa(d.ra.Azi)
    p[keyDtype] = d.DType.String()
    
    w, err := os.Create(path)
    if err != nil {
        return
    }
    defer w.Close()
    
    err = p.Save(w, separator)
    return
}

func (d File) WithShape(dat string, dtype DType) (df File) {
    if dtype == Unknown {
        dtype = d.DType
    }
    
    df = New(dat, d.ra.Rng, d.ra.Azi, dtype)
    return
}


func (d File) Move(dir string) (dm File, err error) {
    var ferr = merr.Make("File.Move")
    
    if dm.Dat, err = Move(d.Dat, dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    dm.ra, dm.DType = d.ra, d.DType
    
    return dm, nil
}

func (d File) Exist() (b bool, err error) {
    var ferr = merr.Make("File.Exist")
    
    if b, err = utils.Exist(d.Dat); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    return b, nil
}

func SameCols(one IFile, two IFile) (err error) {
    var ferr = merr.Make("SameCols")
    
    n1, n2 := one.Rng(), two.Rng()
    
    if n1 != n2 {
        return ferr.Wrap(ShapeMismatchError{
            dat1: one.FilePath(),
            dat2: two.FilePath(),
            n1:n1,
            n2:n2,
            dim: "range samples / columns",
        })
    }
    return nil
}

func SameRows(one IFile, two IFile) error {
    var ferr = merr.Make("SameRows")
    
    n1, n2 := one.Azi(), two.Azi()
    
    if n1 != n2 {
        return ferr.Wrap(ShapeMismatchError{
            dat1: one.FilePath(),
            dat2: two.FilePath(),
            n1:n1,
            n2:n2,
            dim: "azimuth lines / rows",
        })
    }
    
    return nil
}

func SameShape(one IFile, two IFile) (err error) {
    var ferr = merr.Make("SameShape")
    
    if err = SameCols(one, two); err != nil {
        return ferr.Wrap(err)
    }

    if err = SameRows(one, two); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

//func (d *File) Parse() (err error) {
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

func ParseDate(dateStr string) (t time.Time, err error) {
    ferr := merr.Make("ParseDate")
    
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
        //_, err := fun(opt.File, opt.Rng, opt.Start, opt.Nlines, opt.Scale,
                      //opt.Exp)
        
        //if err != nil {
            //return Handle(err, "failed to execute display command")
        //}
    //}
    //return nil
//}


// TODO: implement proper selection of plot command
//func (d File) Raster(opt RasArgs) (err error) {
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
        //_, err = fun(opt.File, opt.Rng, opt.Start, opt.Nlines,
            //opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
            //opt.ImgFmt, opt.HeaderSize, opt.Raster)

    //} else {
        //if len(sec) == 0 {
            //_, err = fun(opt.File, opt.Rng, opt.Start, opt.Nlines,
                //opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                //opt.LR, opt.Raster, opt.ImgFmt, opt.HeaderSize)

        //} else {
            //_, err = fun(opt.File, sec, opt.Rng, opt.Start, opt.Nlines,
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


func (d *File) Set(s string) (err error) {
    *d, err = FromFile(s)
    return
}
