package datafile

// TODO: reimplement so it reads from file instead of reading from
// map
import (
    "fmt"
    "io"
    "strings"
    
    "../utils"
)

type (
    parameters map[string]string
    
    Params struct {
        filePath, sep string
        params parameters
    }
)

func NewParams(filePath, sep string) (p Params, err error) {
    ferr := merr.Make("NewParam")
    
    reader, err := utils.NewReader(filePath)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer reader.Close()
    
    p.params = make(parameters)
    
    for reader.Scan() {
        line := reader.Text()
        
        if len(line) == 0 || !strings.Contains(line, sep) {
            continue
        }
        
        split := strings.Split(line, sep)
        
        if len(split) < 2 {
            err = ferr.Fmt(
                "line '%s' contains separator '%s' but either the " + 
                "key or value is missing", line, sep)
            return
        }
        
        p.params[split[0]] = strings.Trim(split[1], " ")
    }
    
    p.sep = sep
    
    return p, nil
}

func (p Params) Param(name string) (s string, err error) {
    s, ok := p.params[name]
    
    if !ok {
        merr.Make("Params.Param").Wrap(
            ParameterError{
                path: p.filePath,
                par: name,
            })
    }
    
    return s, nil
}

func (p Params) Splitter(name string) (sp utils.SplitParser, err error) {
    ferr := merr.Make("Params.Splitter")
    
    s, err := p.Param(name)
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    sp, err = utils.NewSplitParser(s, " ")
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (p Params) Int(name string, idx int) (i int, err error) {
    ferr := merr.Make("Params.Int")
    
    sp, err := p.Splitter(name)
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    i, err = sp.Int(idx)
    if err != nil {
        err = ferr.Wrap(err)
    }

    return
}

func (p Params) Float(name string, idx int) (f float64, err error) {
    ferr := merr.Make("Params.Float")
    
    sp, err := p.Splitter(name)
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    f, err = sp.Float(idx)
    if err != nil {
        err = ferr.Wrap(err)
    }

    return
}

func (p *Params) Set(key, value string) {
    p.params[key] = value
}

func (p Params) Save(w io.StringWriter) (err error) {
    ferr := merr.Make("Params.Save")
    
    sep := p.sep
    
    for key, val := range p.params {
        s := fmt.Sprintf("%s%s\t%s", key, sep, val)
        
        
        _, err = w.WriteString(s)
        if err != nil {
            return ferr.Wrap(err)
        }
    }
    
    return nil
}

func NewGammaParam(path string) (Params, error) {
    return NewParams(path, ":")
}
