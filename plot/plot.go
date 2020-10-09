package plot

import (
    //"github.com/bozso/gomma/common"
)

type Plottable interface {
    DataDescription() DataDescription
}

type ZeroFlag int

const (
    Missing ZeroFlag = iota
    Valid
)

type Inverse int

const (
    Float2Raster Inverse = 1
    Raster2Float Inverse = -1
)

type Channel int

const (
    Red   Channel = 1
    Green Channel = 2
    Blue  Channel = 3
)

func calcFactor(ndata, factor int) int {
    ret := float64(ndata) / float64(factor)
    
    if ret <= 0.0 {
        return 1
    } else {
        return int(ret)
    }
}
