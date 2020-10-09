package plot

import (
    "github.com/bozso/emath/geometry"
    "github.com/bozso/emath/validate"
    "github.com/bozso/gotoolbox/path"
    
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/settings"
    "github.com/bozso/gomma/data"
)

type DataDescription struct {
    DataFile path.ValidFile
    Type     data.Type
    Mode     Mode    
    common.RngAzi
}

type AverageFactor struct {
    Rng validate.NaturalInt `json:"range"`
    Azi validate.NaturalInt `json:"azimuth"`
}

type Common struct {
    Scale         float64              `json:"scale"`
    Exp           float64              `json:"exp"`
    Start         validate.PositiveInt `json:"start"`
    NumLines      validate.PositiveInt `json:"num_lines"`
    MinMax        geometry.MinMaxFloat `json:"min_max"`    
    HeaderSize    validate.NaturalInt  `json:"header_size,omitempty"`
    Raster        path.Path            `json:"output_raster,omitempty"`
    settings.RasterExtension           `json:"raste_extension,omitempty"`
}

type CommonOptions struct {
    Common
    AverageFactor AverageFactor       `json:"average_factor"`
    Flip          bool                `json:"flip"`
}

func (c CommonOptions) Parse(p Plottable) (o Options) {
    o.Common = c.Common
    o.DataDesc = p.DataDescription()

    dim, avg := &o.DataDesc.RngAzi, &c.AverageFactor
    
    o.AveragePixels.Rng = calcFactor(dim.Rng, int(avg.Rng))
    o.AveragePixels.Azi = calcFactor(dim.Azi, int(avg.Azi))

    if c.Flip {
        o.LR = -1
    } else {
        o.LR = 1
    }
    
    if c.MinMax.Min == 0.0 {
        o.MinMax.Min = 0.1
    }
    
    if c.MinMax.Max == 0.0 {
        o.MinMax.Min = 0.9
    }
    
    return
}

type Options struct {
    AveragePixels common.RngAzi
    MinMax        geometry.MinMaxFloat
    DataDesc      DataDescription
    LR            int
    Common
}

func (o *Options) GetRaster() (p path.Path) {
    if len(o.Raster.String()) == 0 {
        o.Raster = o.DataDesc.DataFile.AddExt(
            string(o.Common.RasterExtension))
    }
    
    return o.Raster
}

type RasterOptions struct {
    Options
}


type DisArgs struct {
    zeroFlag   ZeroFlag
    Sec        string  `name:"sec" default:""`
    StartSec   int     `name:"startSec" default:"1"`
    StartCC    int     `name:"startCC" default:"1"`
    Coh        string  `name:"coh" default:""`
    Cycle      float64 `name:"cycle" default:"160.0"`
    Elev       float64 `name:"elev" default:"45.0"`
    Orient     float64 `name:"orient" default:"135.0"`
    ColPost    float64 `name:"colpost" default:"0"`
    RowPost    float64 `name:"rowpost" default:"0"`
    Offset     float64 `name:"offset" default:"0.0"`
    PhaseScale float64 `name:"scale" default:"0.0"`
    CC         string
    CCMin      float64 `name:"ccMin" default:"0.2"`
}
