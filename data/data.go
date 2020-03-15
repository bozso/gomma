package data

import (
    "fmt"
    "time"
    "strings"
    
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gamma/utils/params"
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/date"
)

type (    
    Pather interface {
        DataPath() path.File
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
        Save(path.File) error
    }
    
    FilePaths struct {
        DatFile   path.File     `json:"datafile"`
        ParFile   path.File     `json:"parameterfile"`        
    }
    
    File struct {
        Paths
        Dtype     Type          `json:"data_type"`
        Ra        common.RngAzi `json:"rng_azi"`
        time.Time
    }
)

func NewGammaParams(file path.File) (p params.Params, err error) {
    return params.FromFile(file, separator)
}

func (f File) JsonName() (file path.File) {
    // it is okay for the path to not exist
    file, _ = path.New(fmt.Sprintf("%s.json", f.DatFile)).ToFile()
    return
}

func (d File) Rng() int {
    return d.Ra.Rng
}

func (d File) Azi() int {
    return d.Ra.Azi
}

func (d File) DataPath() path.File {
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
    return d.SaveWithPath(d.JsonName())
}

func (d File) SaveWithPath(file path.File) (err error) {
    return common.SaveJson(file, d)
}

func (d File) WithShape(dat path.File, dtype Type) File {
    if dtype == Unknown {
        dtype = d.Dtype
    }
    
    return File{
        Paths: d.Paths,
        Ra: d.Ra,
        Dtype: dtype,
    }
}

func (d File) Move(dir path.Dir) (dm File, err error) {
    dm = d
    p, err := d.DatFile.Move(dir)
    if err != nil {
        return
    }
    
    dm.DatFile, err = p.ToFile()
    if err != nil {
        return
    }
    
    par := d.ParFile
    
    exist, err := par.Exist()
    if err != nil {
        return
    }
    
    if exist {
        p, err = par.Move(dir)
        if err != nil {
            return
        }
    }
    
    dm.ParFile, err = p.ToFile()
    if err != nil {
        return
    }
    
    
    dm.Ra, dm.Dtype = d.Ra, d.Dtype
    
    return
}

func (d File) Exist() (b bool, err error) {
    b, err = d.DatFile.Exist()
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
    dat1, dat2 path.File
    dim string
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
    file, err := path.New(s).ToFile()
    if err != nil {
        return
    }
    
    
    err = common.LoadJson(file, d)
    return
}

const (
    separator = ":"
)
