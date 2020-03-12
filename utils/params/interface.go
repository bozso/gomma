package params

import (
    "fmt"
    "errors"

    "github.com/bozso/gotoolbox/splitted"
)

type Inter interface {
    Int(key string, idx int) (int, error)
}

type Floater interface {
    Float(key string, idx int) (float64, error)
}

type Retreiver interface {
    Param(key string) (string, error)
}

/*
Wrapper for interface Retreiver. If something implements the
Retreiver interface, wrapping it in this struct will automatically
implement the Inter and Floater interfaces.
*/
type Parser struct {
    Retreiver
}

func (p Parser) Splitter(key string) (sp splitted.Parser, err error) {
    s, err := p.Param(key)
    if err != nil {
        return
    }
    
    sp, err = splitted.New(s, " ")
    return
}

func (p Parser) Int(key string, idx int) (ii int, err error) {
    sp, err := p.Splitter(key)
    if err != nil {
        return
    }
    
    ii, err = sp.Int(idx)
    return
}

func (p Parser) Float(key string, idx int) (ff float64, err error) {
    sp, err := p.Splitter(key)
    if err != nil {
        return
    }
    
    ff, err = sp.Float(idx)
    return
}

/*
Wrapper for reading from many Retreivers. It will try to read
the appropriate parameters from each of the receivers. Returns
with error if parameter could not be found in any of them.
*/
type TeeParser struct {
    retreivers []Retreiver
}

// Wrap it into a parser struct.
func (t TeeParser) ToParser() (p Parser) {
    return Parser{t}
}

func NewTeeParser(retreiver ...Retreiver) (t TeeParser) {
    t.retreivers = retreiver
    return
}


var nf = NotFound{}

func (tp TeeParser) Param(key string) (s string, err error) {
    for _, r := range tp.retreivers {
        s, err = r.Param(key)
        
        if err == nil {
            return
        }
        
        // try to retreive the parameter in the next retreiver
        if errors.Is(err, nf) {
            continue
        } else {
            return
        }
    }
    
    if len(s) == 0 {
        err = fmt.Errorf(
            "parameter '%s' was not found in any of the retreivers",
            key)
    }
    
    return
}
