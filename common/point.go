package common

import (
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/emath/geometry"

    "github.com/bozso/gomma/utils/params"
)

type AOI [4]geometry.LatLon

type MinOrMax int

const (
    Min MinOrMax = iota
    Max
)

type LatLonRegion struct {
    Max geometry.LatLon `json:"max"`
    Min geometry.LatLon `json:"min"`
}

func (ll LatLonRegion) Contains(l geometry.LatLon) (b bool) {
    return ll.ToRegion().Contains(l.ToPoint())
}

func (ll LatLonRegion) ToRegion() (r geometry.Region) {
    r.Max, r.Min = ll.Max.ToPoint(), ll.Min.ToPoint()
    return
}

func ParseRegion(info params.Parser) (r LatLonRegion, err error) {
    r.Max, err = Max.Parse(info)
    if err != nil {
        return
    }

    r.Min, err = Min.Parse(info)
    return
}

func (mode MinOrMax) Parse(info params.Parser) (p geometry.LatLon, err error) {
    var tpl_lon, tpl_lat string

    switch mode {
    case Max:
        tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
    case Min:
        tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
    }

    if p.Lon, err = info.Float(tpl_lon, 0); err != nil {
        err = ParseError{"longitude", err}
        return
    }

    if p.Lat, err = info.Float(tpl_lat, 0); err != nil {
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
