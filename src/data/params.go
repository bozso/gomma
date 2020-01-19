package data

// TODO: reimplement so it reads from file instead of reading from
// map
import (
    "fmt"
    "io"
    "strings"
    
    "../utils"
)

type params map[string]string

func fromFile(path, sep string) (p params, err error) {
    reader, err := utils.NewReader(path)
    if err != nil {
        return
    }
    defer reader.Close()
    
    p = make(params)
    
    for reader.Scan() {
        line := reader.Text()
        
        if len(line) == 0 || !strings.Contains(line, sep) {
            continue
        }
        
        split := strings.Split(line, sep)
        
        if len(split) < 2 {
            err = fmt.Errorf(
                "line '%s' contains separator '%s' but either the " + 
                "key or value is missing", line, sep)
            return
        }
        
        p[split[0]] = strings.Trim(split[1], " ")
    }
    
    return p, nil
}

func (p params) Param(name string) (s string, err error) {
    err = nil
    s, ok := p[name]
    
    if !ok {
        err = KeyError{key:name}
    }
    
    return
}

func (p params) Splitter(name string) (sp utils.SplitParser, err error) {
    s, err := p.Param(name)
    if err != nil {
        return
    }
    
    sp, err = utils.NewSplitParser(s, " ")
    return
}

func (p params) Int(name string, idx int) (i int, err error) {
    sp, err := p.Splitter(name)
    if err != nil {
        return
    }
    
    i, err = sp.Int(idx)
    return
}

func (p params) Float(name string, idx int) (f float64, err error) {
    sp, err := p.Splitter(name)
    if err != nil {
        return
    }
    
    f, err = sp.Float(idx)
    return
}

func (p params) Save(w io.StringWriter, sep string) (err error) {
    for key, val := range p {
        s := fmt.Sprintf("%s%s\t%s", key, sep, val)
        
        _, err = w.WriteString(s)
        if err != nil {
            return
        }
    }
    return nil
}

type ParamReader struct {
    path string
    err error
    params
}

func NewReader(path, sep string) (p ParamReader) {
    p.params, p.err = fromFile(path, sep)
    return
}

func (p *ParamReader) wrap(name string) {
    p.err = utils.WrapFmt(p.err,
            "failed to retreive paramter '%s' from file '%s'",
            name, p.path) 
}

func (p *ParamReader) Param(name string) (s string) {
    if p.err != nil {
        return
    }
    
    s, p.err = p.params.Param(name)
    
    if p.err != nil {
        p.wrap(name)
    }
    
    return
}

func (p *ParamReader) Int(name string, idx int) (ii int) {
    if p.err != nil {
        return
    }
    
    ii, p.err = p.params.Int(name, idx)
    
    if p.err != nil {
        p.wrap(name)
    }
    
    return
}

func (p *ParamReader) Float(name string, idx int) (ff float64) {
    if p.err != nil {
        return
    }
    
    ff, p.err = p.params.Float(name, idx)
    
    if p.err != nil {
        p.wrap(name)
    }
    
    return
}

func (p ParamReader) Wrap() (err error) {
    err = p.err
    
    if err != nil {
        err = utils.WrapFmt(err,
            "failed while reading parameters from file '%s'", p.path)
    }
    
    return err
}

type Params struct {
    filePath, sep string
    params
}

func NewParams(path, sep string) (p Params, err error) {
    p.filePath, p.sep = path, sep
    
    p.params, err = fromFile(path, sep)
    return
}

func (p Params) Save(w io.StringWriter) (err error) {
    p.params.Save(w, p.sep)
    return
}

func NewGammaParam(path string) (Params, error) {
    return NewParams(path, ":")
}
