package params

// TODO: reimplement so it reads from file instead of reading from
// map
import (
    "fmt"
    "os"
    "io"
    "strings"
    
    "github.com/bozso/gamma/utils"
)

const defaultLen = 15

type (
    Key string
    Separator string
    
    params []string    
    
    Params struct {
        path, sep string
        p params
    }
)

func New(path, sep string) (p Params) {
    p.path, p.sep = path, sep
    
    p.p = make(params, 1)
    
    return
}

func WithLen(path, sep string, count int) (p Params) {
    p.path, p.sep = path, sep
    
    p.p = make(params, count)
    
    return
}

func FromFile(path, sep string) (p Params, err error) {
    p = New(path, sep)
    
    reader, err := utils.NewReader(path)
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
    
    return p, nil
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

    err = ParameterError{key:key, path:p.path}
    return
}

func (p Params) Splitter(key string) (sp utils.SplitParser, err error) {
    s, err := p.Param(key)
    if err != nil {
        return
    }
    
    sp, err = utils.NewSplitParser(s, " ")
    return
}


func (p Params) Int(key string, idx int) (ii int, err error) {
    sp, err := p.Splitter(key)
    if err != nil {
        return
    }
    
    ii, err = sp.Int(idx)
    return
}

func (p Params) Float(key string, idx int) (ff float64, err error) {
    sp, err := p.Splitter(key)
    if err != nil {
        return
    }
    
    ff, err = sp.Float(idx)
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
    path := p.path
    
    w, err := os.Create(path)
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

type ParameterError struct {
    path, key string
    err error
}

func (p ParameterError) Error() string {
    return fmt.Sprintf("failed to retreive parameter '%s' from file '%s'",
        p.key, p.path)
}

func (p ParameterError) Unwrap() error {
    return p.err
}

var ParamError = ParameterError{}
