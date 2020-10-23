package service

import (
    "github.com/bozso/gomma/mli"
    "github.com/bozso/gomma/slc"
    "github.com/bozso/gomma/dem"
    "github.com/bozso/gomma/geo"
)

var plottables = [...]Plottable{
    mli.MLI{},
    mli.SLC{},
    dem.File{},
    geo.Height{},
}
