package plot

import (
    "github.com/bozso/gotoolbox/command"    
    //"github.com/bozso/gomma/settings"
)

type Plotter interface {
    Raster(Plottable, Options) error
    Display(Plottable, Options) error
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
        c.Raster, c.Display = "rasbyte", "disbyte"
    }
    
    return
}

func PlotByte(cmd command.Command, o Options) (err error) {
    dd := &o.DataDesc
    _, err = cmd.Call(dd.DataFile, dd.Rng, o.Start, o.NumLines,
                     o.AveragePixels.Rng, o.AveragePixels.Azi, o.Scale,
                     o.LR, o.GetRaster())
    return
}
