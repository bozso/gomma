package plot

import (
    "github.com/bozso/gotoolbox/errors"

)

type Service interface {
    Plot(DataDescription, Type, Options) error
}

type DefaultService struct {
    plotters map[Mode]Plotter
}

func (d DefaultService) Plot(dd DataDescription, t Type, op Options) (err error) {
    pl, ok := d.plotters[dd.Mode]
    if !ok {
        err = errors.KeyNotFound(dd.Mode.String())
        return
    }


    return
}
