package service

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/command"

    //"github.com/bozso/gomma/settings"
    "github.com/bozso/gomma/plot"
)

type Service interface {
    Raster(path.ValidFile, plot.CommonOptions) error
    Display(path.ValidFile, plot.CommonOptions) error
}

type Commands [MaximumMode]command.Command

type ServiceImpl struct {
    plotters [MaximumMode]plot.Plotter
}

func (s *ServiceImpl) Plot(t Type, p Plottable, co plot.CommonOptions) (err error) {
    opt := co.Parse(p)
    return plotters[opt.Mode].Plot(t, opt)
}
