package common

import (
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/geometry"

    "github.com/bozso/gomma/utils/params"
)

type (
    Point struct {
        X, Y float64
    }
    
    AOI [4]Point
)

func (p Point) InRectangle(r Rectangle) bool {
    return (p.X < r.Max.X && p.X > r.Min.X &&
            p.Y < r.Max.Y && p.Y > r.Min.Y)
}

type MinOrMax int

const (
    Min MinOrMax = iota
    Max
)

func (mode MinOrMax) ParsePoint(info params.Parser) (p Point, err error) {
    var tpl_lon, tpl_lat string
    
    switch mode {
    case Max:
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    case Min:
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if p.X, err = info.Float(tpl_lon, 0); err != nil {
        err = ParseError{"longitude", err}
        return
    }

    if p.Y, err = info.Float(tpl_lat, 0); err != nil {
        err = ParseError{"latitude", err}
        return
    }

    return
}

type ParseError struct {
    coordinate string
    err error
}

func (e ParseError) Error() string {
    const msg errors.String = "failed to retreive %s value"
    
    return msg.WrapFmt(e.err, e.coordinate).Error()
}

func (e ParseError) Unwrap() error {
    return e.err
}
