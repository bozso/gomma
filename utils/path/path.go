package path

import (
    "os"
    "path/filepath"
    "io/ioutil"
    
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/stream"
)

type Joiner struct {
    s string
}

func NewJoiner(p string) (j Joiner) {
    j.s = p
    return
}

func (j Joiner) Join(elems ...string) (s string) {
    ss := []string{j.s}
    
    return filepath.Join(append(ss, elems...)...)
}

func Exist(s string) (b bool, err error) {
    b = false
    _, err = os.Stat(s)

    if err == nil {
        b = true
        return
    }
    
    if os.IsNotExist(err) {
        err = nil
        return
    }
    
    err = utils.WrapFmt(err, "failed to check wether file '%s' exists", s)
    return
}
func Move(path string, dir string) (s string, err error) {
    dst, err := filepath.Abs(filepath.Join(dir, filepath.Base(path)))
    if err != nil {
        err = utils.WrapFmt(err, "failed to create absolute path")
        return
    }
    
    if err = os.Rename(path, dst); err != nil {
        return
    }
    
    return dst, nil
}

func Mkdir(name string) (err error) {
    if err = os.MkdirAll(name, os.ModePerm); err != nil {
        err = utils.WrapFmt(err, "failed to create directory '%s'", name)
    }
    
    return
}


type Path struct {
    s string
}

func (p Path) String() string {
    return p.s
}

func (p Path) Abs() (pp Path, err error) {
    pp.s, err = filepath.Abs(p.s)
    return
}

func (p Path) Len() int {
    return len(p.s)
}

func ReadFile(p string) (b []byte, err error) {
    var f *stream.In
    if err = f.Set(p); err != nil {
        return
    }
    defer f.Close()
    
    b, err = ioutil.ReadAll(f)
    return
}


/*
type Files []*File

func (f Files) String() string {
    if f != nil {
        // TODO: list something sensible
        return ""
    }
    
    return ""
}

func (f Files) Set(s string) (err error) {
    split := strings.Split(s, ",")
    
    f = make(Files, len(split))
    
    for ii, fpath := range f {
        if err = fpath.Set(split[ii]); err != nil {
            return
        }
    }
    return nil
}
*/
