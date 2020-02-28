package common

import (
    "fmt"

    "github.com/bozso/gamma/utils"
)

type LatLon struct {
    Lat float64 `name:"lan" default:"1.0"`
    Lon float64 `name:"lot" default:"1.0"`
}

func (ll LatLon) String() string {
    return fmt.Sprintf("%f,%f", ll.Lon, ll.Lat)
}

func (ll *LatLon) Set(s string) (err error) {
    if len(s) == 0 {
        return utils.EmptyStringError{}
    }
    
    split, err := utils.NewSplitParser(s, ",")
    if err != nil {
        return
    }
    
    ll.Lat, err = split.Float(0)
    if err != nil {
        return
    }

    ll.Lon, err = split.Float(1)
    if err != nil {
        return
    }
    
    return nil
}

type Point struct {
    X, Y float64
}

func (p Point) InRect(r Rectangle) bool {
    return (p.X < r.Max.X && p.X > r.Min.X &&
            p.Y < r.Max.Y && p.Y > r.Min.Y)
}

type Rectangle struct {
    Max, Min Point
}

type AOI [4]Point

