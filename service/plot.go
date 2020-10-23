package service

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/command"
    
    //"github.com/bozso/gomma/settings"
)

type Service interface {
    Raster(path.ValidFile, CommonOptions) error
    Display(path.ValidFile, CommonOptions) error
}

type Commands [MaximumMode]command.Command

type ServiceImpl struct {
    plotters [MaximumMode]Plotter
}

func (s *ServiceImpl) Plot(t Type, vf path.ValidFile, co CommonOptions) (err error) {
    bytes, err := vf.ReadAll()
    if err != nil {
        return
    }
    
    opt := co.Parse(datafile)
    
    return plotters[opt.Mode].Plot(t, opt)    
}
