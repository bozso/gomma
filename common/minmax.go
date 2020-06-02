package common

import (
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/splitted"
)

type(
    Minmax struct {
        Min float64 `name:"min" default:"0.0"`
        Max float64 `name:"max" default:"1.0"`
    }
    
    IMinmax struct {
        Min int `name:"min" default:"0"`
        Max int `name:"max" default:"1"`
    }
)

func (mm *IMinmax) Set(s string) (err error) {
    const field errors.NotEmpty = "IMinmax"
    
    if err = field.Check(s); err != nil {
        return
    }
    
    split, err := splitted.New(s, ",")
    if err != nil {
        return
    }
    
    mm.Min, err = split.Int(0)
    if err != nil {
        return
    }
    
    mm.Max, err = split.Int(1)
    if err != nil {
        return
    }
    
    return nil
}
