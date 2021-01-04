package plot

import (

)

type BytePlotter PlotCommand

func (b BytePlotter) Plot(t Type, o Options) (err error) {
    m := &o.Meta

    switch t {
    case Raster:
    _, err = b.raster.Call(m.DataFile, m.RngAzi.Rng, o.Start, o.NumLines,
                     o.AveragePixels.Rng, o.AveragePixels.Azi, o.Scale,
                     o.LR, o.GetRaster())
    /// TODO: check for arguments
    case Display:
    _, err = b.display.Call(m.DataFile, m.RngAzi.Rng, o.Start, o.NumLines,
                     o.AveragePixels.Rng, o.AveragePixels.Azi, o.Scale,
                     o.LR, o.GetRaster())
    }
    return
}
