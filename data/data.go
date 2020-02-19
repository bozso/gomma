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
    "github.com/bozso/gamma/date"
)

type (    
    Pather interface {
        DataPath() string
    }
    
    Typer interface {
        DataType() Type
    }
    
    File struct {
        DatFile   string     `json:"datafile"`
        ParFile   string     `json:"parameterfile"`
        Dtype Type           `json:"data_type"`
        Ra    common.RngAzi
        time.Time
    }
)

func newGammaParams(path string) (p params.Params, err error) {
    return params.FromFile(path, separator)
}

func Import(path, rng, azi, Type string) (f File, err error) {
    f.ParFile = path
    
    
}

func FromFile(path string) (d File, err error) {
    d.Par = path
    
    pr, err := newGammaParams(path)
    if err != nil { return }
    
    d.Dat, err = pr.Param(keyDatafile)
    if err != nil { return }
    
    d.Ra.Rng, err = pr.Int(keyRng, 0)
    if err != nil { return }
    
    d.Ra.Azi, err = pr.Int(keyAzi, 0)
    if err != nil { return }
    
    ds, err := pr.Param(keyDtype)
    if err != nil { return }
    
    err = d.Dtype.Set(ds)

    ds, err = pr.Param(keyDate)

    if err != nil {
        if errors.Is(err, params.ParamError) {
            return
        }
        
        d.Time, err = DateFmt.Parse(ds)
    }
    
    return
}

func (d File) Rng() int {
    return d.Ra.Rng
}

func (d File) Azi() int {
    return d.Ra.Azi
}

func (d File) DataPath() string {
    return d.Dat
}

//func (d File) Date() time.Time {
    //return d.time
//}

func (d File) DataType() Type {
    return d.Dtype
}

func (d File) TypeCheck(ftype, expect string, dtypes... Type) (err error) {
    D := d.Dtype
    
    for _, dt := range dtypes {
        if D == dt {
            return nil
        }
    }
    
    return TypeMismatchError{ftype:ftype, expected:expect, Type:D}
}

func (d File) Save() (err error) {
    path := d.Par
    
    exists, err := utils.Exist(path)
    if err != nil { return; }
    
    var p params.Params
    if exists {
        p, err = params.FromFile(path, separator)
        if err != nil { return }
    } else {
        p = params.New(path, separator)
    }
    
    p.SetVal(keyDatafile, d.Dat)
    p.SetVal(keyRng, strconv.Itoa(d.Ra.Rng))
    p.SetVal(keyAzi, strconv.Itoa(d.Ra.Azi))
    p.SetVal(keyDtype, d.Dtype.String())
    
    err = p.Save()
    return
}

func (d File) WithShape(dat string, dtype Type) (df File) {
    if dtype == Unknown {
        dtype = d.Dtype
    }
    
    df = New(dat, d.Ra.Rng, d.Ra.Azi, dtype)
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
    
    dm.Ra, dm.Dtype = d.Ra, d.Dtype
    
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
            dat1: one.DataPath(),
            dat2: two.DataPath(),
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
            dat1: one.DataPath(),
            dat2: two.DataPath(),
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

const DateFmt date.ParseFmt = "2016 12 05"

type(
    Subset struct {
        RngOffset int
        AziOffset int
        RngWidth int
        AziLines int
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
