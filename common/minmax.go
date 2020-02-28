package common

import (
    "github.com/bozso/gamma/utils"
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
    if len(s) == 0 {
        return utils.EmptyStringError{}
    }
    
    split, err := utils.NewSplitParser(s, ",")
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
