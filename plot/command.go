package plot

import (
    "github.com/bozso/gotoolbox/command"    
    "github.com/bozso/gomma/settings"    
)

type Plotter interface {
    Raster(Plottable, RasArgs) error
    Display(Plottable, DisArgs) error
}

type CommandNames struct {
    Raster, Display string
}

type CommandNamer interface {
    CommandNames() CommandNames
}

func (m Mode) CommandNames() (c CommandNames) {
    switch m {
    case Byte:
        raster, display = "rasbyte", "disbyte"
    }
    
    c.raster, c.display = raster, display
    return
}

func PlotByte(cmd command.Command, opt RasArgs) (err error) {
    _, err = cmd.Call(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                     opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.LR,
                     opt.Raster)
    return
}
