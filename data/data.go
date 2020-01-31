package data

// TODO: seperate field for storing rng, azi, DType values

import (
    "fmt"
    "strconv"
    "errors"
    "time"
    
    
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/params"
    "github.com/bozso/gamma/common"
    //"../utils"
    //"../utils/params"
    //"../common"
)

const merr = utils.ModuleName("gamma.datafile")

const (
    keyDatafile = "golang_meta_datafile"
    keyRng = "golang_meta_rng"
    keyAzi = "golang_meta_azi"
    keyDtype = "golang_meta_dtype"
    keyDate = "date"
    separator = ":"
)

var emptyTime = time.Time{}

type (    
    IFile interface {
        FilePath() string
        Rng() int
        Azi() int
        DataType() Type
        //Move(string) (File, error)
    }
    
    OptTime struct {
        time.Time
        present bool
    }
    
    File struct {
        Dat, Par string
        ra       common.RngAzi
        time     OptTime
        dtype    Type
    }
)

func New(dat string, rng, azi int, dtype Type) (d File) {
    d.Dat, d.Par = dat, dat + ".par"
    d.ra.Rng, d.ra.Azi, d.dtype = rng, azi, dtype
    
    return
}

func newGammaParams(path string) (p params.Params, err error) {
    return params.FromFile(path, separator)
}

func FromFile(path string) (d File, err error) {
    d.Par = path
    
    pr, err := newGammaParams(path)
    if err != nil { return }
    
    d.Dat, err = pr.Param(keyDatafile)
    if err != nil { return }
    
    d.ra.Rng, err = pr.Int(keyRng, 0)
    if err != nil { return }
    
    d.ra.Azi, err = pr.Int(keyAzi, 0)
    if err != nil { return }
    
    ds, err := pr.Param(keyDtype)
    if err != nil { return }
    
    err = d.dtype.Set(ds)

    ds, err = pr.Param(keyDate)
    if err == nil {
        t, err := DateFmt.Parse(ds)
        if err != nil { return d, err }
        
        d.time = OptTime{t, true}
    }
    
    if errors.Is(err, params.ParamError) {
        d.time = OptTime{present:false}
    }
    
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

//func (d File) Date() time.Time {
    //return d.time
//}

func (d File) DataType() Type {
    return d.dtype
}

func (d File) TypeCheck(ftype, expect string, dtypes... Type) (err error) {
    b, D := false, d.dtype
    
    for _, dt := range dtypes {
        if D == dt {
            b = true
            break
        }
    }
    
    if !b {
        err = TypeMismatchError{ftype:ftype, expected:expect, Type:D}
        return
    }
    
    return nil
}

func (d File) Save() (err error) {
    path := d.Par
    
    exists, err := utils.Exist(path)
    if err != nil { return; }
    
    var p params.Params
    if exists {
        p, err = params.FromFile(path, separator)
        if err != nil { return; }
    } else {
        p = params.New(path, separator)
    }
    
    p.SetVal(keyDatafile, d.Dat)
    p.SetVal(keyRng, strconv.Itoa(d.ra.Rng))
    p.SetVal(keyAzi, strconv.Itoa(d.ra.Azi))
    p.SetVal(keyDtype, d.dtype.String())
    
    err = p.Save()
    return
}

func (d File) WithShape(dat string, dtype Type) (df File) {
    if dtype == Unknown {
        dtype = d.dtype
    }
    
    df = New(dat, d.ra.Rng, d.ra.Azi, dtype)
    return
}


func (d File) Move(dir string) (dm File, err error) {
    dm.Dat, err = utils.Move(d.Dat, dir)
    if err != nil {
        return
    }

    dm.Par, err = utils.Move(d.Par, dir)
    if err != nil {
        return
    }
    
    dm.ra, dm.dtype = d.ra, d.dtype
    
    return
}

func (d File) Exist() (b bool, err error) {
    b, err = utils.Exist(d.Dat)
    return
}

func SameCols(one IFile, two IFile) (err error) {
    n1, n2 := one.Rng(), two.Rng()
    
    if n1 != n2 {
        return ShapeMismatchError{
            dat1: one.FilePath(),
            dat2: two.FilePath(),
            n1:n1,
            n2:n2,
            dim: "range samples / columns",
        }
    }
    return nil
}

func SameRows(one IFile, two IFile) error {
    n1, n2 := one.Azi(), two.Azi()
    
    if n1 != n2 {
        return ShapeMismatchError{
            dat1: one.FilePath(),
            dat2: two.FilePath(),
            n1:n1,
            n2:n2,
            dim: "azimuth lines / rows",
        }
    }
    
    return nil
}

func SameShape(one IFile, two IFile) (err error) {
    
    err = SameCols(one, two)
    if err != nil {
        return
    }

    err = SameRows(one, two)

    return
}

type ShapeMismatchError struct {
    dat1, dat2, dim string
    n1, n2 int
    err error
}

func (s ShapeMismatchError) Error() string {
    return fmt.Sprintf("expected datafile '%s' to have the same %s as " + 
                       "datafile '%s' (%d != %d)", s.dat1, s.dim, s.dat2, s.n1,
                       s.n2)
}

func (s ShapeMismatchError) Unwrap() error {
    return s.err
}


const DateFmt common.DateFormat = "2016 12 05"

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



func (d *File) Set(s string) (err error) {
    *d, err = FromFile(s)
    return
}
