package data

import (
    "fmt"
    "time"
    "strings"
    
    
    "github.com/bozso/gamma/utils/params"
    "github.com/bozso/gamma/utils/path"
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
    
    Data interface {
        Typer
        Pather
        common.Dims
    }
    
    Saver interface {
        Save(string) error
    }
    
    File struct {
        DatFile   string        `json:"datafile"`
        ParFile   string        `json:"parameterfile"`
        Dtype     Type          `json:"data_type"`
        Ra        common.RngAzi `json:"rng_azi"`
        time.Time
    }
)

func newGammaParams(path string) (p params.Params, err error) {
    return params.FromFile(path, separator)
}

func FromDataPath(p, ext string) (f File) {
    if len(ext) == 0 {
        ext = "par"
    }
    
    f.DatFile = p
    f.ParFile = fmt.Sprintf("%s.%s", p, ext)
    return
}

func (f File) JsonName() string {
    return fmt.Sprintf("%s.json", f.DatFile)
}

func (d File) Rng() int {
    return d.Ra.Rng
}

func (d File) Azi() int {
    return d.Ra.Azi
}

func (d File) DataPath() string {
    return d.DatFile
}

func (d File) Date() time.Time {
    return d.Time
}

func (d File) DataType() Type {
    return d.Dtype
}

func (d File) TypeCheck(dtypes... Type) (err error) {
    D := d.Dtype
    
    for _, dt := range dtypes {
        if D == dt {
            return nil
        }
    }
    
    sb := strings.Builder{}
    
    for _, dt := range dtypes {
        sb.WriteString(dt.String() + ", ")
    }
    
    return TypeMismatchError{
        datafile:d.DatFile,
        expected:sb.String(),
        Type:D,
    }
}

func (d File) Save(p string) (err error) {
    if len(p) == 0 {
        p = d.JsonName()
    }
    
    return common.SaveJson(p, d)
}

func (d File) WithShape(dat string, dtype Type) File {
    if dtype == Unknown {
        dtype = d.Dtype
    }
    
    return File{
        DatFile: dat,
        Ra: d.Ra,
        Dtype: dtype,
    }
}

func (f File) GetParser() (p params.Params, err error) {
    p, err = newGammaParams(f.ParFile)
    return
}

func (f *File) Load(k *ParamKeys) (err error) {
    p, err := f.GetParser()
    if err != nil { return }
    
    if k == nil {
        k = &DefaultKeys
    }
    
    return f.LoadWithParser(k, params.Parser{p})
}

func (f *File) LoadWithParser(k *ParamKeys, pr params.Parser) (err error) {
    if err != nil { return }
    
    f.Ra.Rng, err = pr.Int(k.RngKey, 0)
    if err != nil { return }
    
    f.Ra.Azi, err = pr.Int(k.AziKey, 0)
    if err != nil { return }
    
    s, err := pr.Param(k.TypeKey)
    if err != nil { return }
    
    err = f.Dtype.Set(s)
    if err != nil { return }
    
    if d := k.DateKey; len(d) != 0 {
        s, err = pr.Param(d)
        if err != nil { return }
        
        f.Time, err = DateFmt.Parse(s)
    }
    
    return
}

func (d File) Move(dir string) (dm File, err error) {
    dm.DatFile, err = path.Move(d.DatFile, dir)
    if err != nil { return }
    
    p := d.ParFile
    
    if len(p) > 0 {
        p, err = path.Move(p, dir)
        if err != nil { return }
    }
    dm.ParFile = p
    
    dm.Ra, dm.Dtype = d.Ra, d.Dtype
    
    return
}

func (d File) Exist() (b bool, err error) {
    b, err = path.Exist(d.DatFile)
    return
}

func SameCols(one common.Dims, two common.Dims) *ShapeMismatchError {
    n1, n2 := one.Rng(), two.Rng()
    
    if n1 != n2 {
        return &ShapeMismatchError{
            n1:n1,
            n2:n2,
            dim: "range samples / columns",
        }
    }
    return nil
}

func SameRows(one common.Dims, two common.Dims) *ShapeMismatchError {
    n1, n2 := one.Azi(), two.Azi()
    
    if n1 != n2 {
        return &ShapeMismatchError{
            n1:n1,
            n2:n2,
            dim: "azimuth lines / rows",
        }
    }
    
    return nil
}

func SameShape(one common.Dims, two common.Dims) (err *ShapeMismatchError) {
    err = SameCols(one, two)
    if err != nil { return }

    return SameRows(one, two)
}

type ShapeMismatchError struct {
    dat1, dat2, dim string
    n1, n2 int
    err error
}

func (s ShapeMismatchError) Error() string {
    return fmt.Sprintf("expected datafile '%s' to have the same %s as " + 
        "datafile '%s' (%d != %d)", s.dat1, s.dim, s.dat2, s.n1, s.n2)
}

func (s ShapeMismatchError) Pathes(one, two Pather) error {
    s.dat1, s.dat2 = one.DataPath(), two.DataPath()
    
    return s
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
    err = common.LoadJson(s, d)
    return
}

const (
    separator = ":"
)
