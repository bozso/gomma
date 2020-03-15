package params

// TODO: reimplement so it reads from file instead of reading from
// map
import (
    "fmt"
    "os"
    "io"
    "strings"

    "github.com/bozso/gotoolbox/path"
)

const defaultLen = 15

type (
    Key string
    Separator string
    
    params []string
    
    Params struct {
        file path.Path
        sep string
        p params
    }
)

func (p Params) ToParser() (par Parser) {
    par.Retreiver = p
    return
}

func New(file path.Path, sep string) (p Params) {
    p = WithLen(file, sep, 1)
    return
}

func WithLen(file path.Path, sep string, count int) (p Params) {
    p.file, p.sep = file, sep
    
    p.p = make(params, count)
    
    return
}

func FromFile(file path.Path, sep string) (p Params, err error) {
    p = New(file, sep)
    
    f, err := file.ToFile()
    
    reader, err := f.Scanner()
    if err != nil {
        return
    }
    defer reader.Close()
    
    for reader.Scan() {
        line := reader.Text()
        
        if len(line) == 0 || !strings.Contains(line, sep) {
            continue
        }
        
        p.p = append(p.p, line)
    }
    
    return
}

func FromString(elems, sep string) (p Params) {
    p.sep = sep
    split := strings.Split(elems, "\n")
    
    for _, line := range split {
        if len(line) == 0 || !strings.Contains(line, sep) {
            continue
        }
        
        p.p = append(p.p, line)
    }
    
    return
}

func (p Params) Param(key string) (s string, err error) {
    sep := p.sep
    
    for _, line := range p.p {
        if !strings.Contains(line, key) {
            continue
        }
        
        split := strings.Split(line, sep)
        
        if len(split) < 2 {
            err = fmt.Errorf(
                "line '%s' contains separator '%s' but either the " + 
                "key or value is missing", line, sep)
            return
        }
        
        s = strings.Trim(split[1], " ")
        return
    }

    err = NotFound{key:key, path:p.file.String()}
    return
}


func (p Params) SetVal(key, val string) {
    s := fmt.Sprintf("%s%s%s", key, p.sep, val)
    
    for ii, line := range p.p {
        if strings.Contains(line, key) {
            p.p[ii] = s
            return
        }
    }
    
    p.p = append(p.p, s)
}

func (p Params) Save() (err error) {
    file, err := p.file.ToFile()
    if err != nil {
        return
    }
    
    w, err := file.Create()
    if err != nil {
        return
    }
    defer w.Close()
    
    err = p.SaveTo(w)
    return
}

func (p Params) SaveTo(w io.StringWriter) (err error) {
    _, err = w.WriteString(strings.Join([]string(p.p), "\n"))
    return
}

type NotFound struct {
    path, key string
    err error
}

func (p NotFound) Error() string {
    return fmt.Sprintf("failed to retreive parameter '%s' from file '%s'",
        p.key, p.path)
}

func (p NotFound) Unwrap() error {
    return p.err
}
