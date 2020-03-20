package common

import (
    "github.com/bozso/gomma/utils/params"
)

type Rectangle struct {
    Max, Min Point
}

func ParseRectangle(info params.Parser) (r Rectangle, err error) {
    r.Max, err = Max.ParsePoint(info)
    if err != nil {
        return
    }
    
    r.Min, err = Min.ParsePoint(info)
    return
}
