package plot

import (
    "github.com/bozso/gamma/common"
    "github.com/bozso/gamma/data"
)

type DisArgs struct {
    ScaleExp
    common.RngAzi
    common.Minmax
    data.Type
    Inverse
    Channel
    Mode       PlotMode
    zeroFlag   ZeroFlag
    Flip       bool    `name:"flip" default:""`
    Datfile    string  `name:"dat" default:""`
    Start      int     `name:"start" default:"0"`
    Nlines     int     `name:"nlines" default:"0"`
    Sec        string  `name:"sec" default:""`
    StartSec   int     `name:"startSec" default:"1"`
    StartCC    int     `name:"startCC" default:"1"`
    Coh        string  `name:"coh" default:""`
    Cycle      float64 `name:"cycle" default:"160.0"`
    LR         int
    Elev       float64 `name:"elev" default:"45.0"`
    Orient     float64 `name:"orient" default:"135.0"`
    ColPost    float64 `name:"colpost" default:"0"`
    RowPost    float64 `name:"rowpost" default:"0"`
    Offset     float64 `name:"offset" default:"0.0"`
    PhaseScale float64 `name:"scale" default:"0.0"`
    CC         string
    CCMin      float64 `name:"ccMin" default:"0.2"`
}

func (arg *DisArgs) Parse(dat data.IFile) {
    arg.ScaleExp.Parse()
    
    if arg.Start == 0 {
        arg.Start = 1
    }
    
    if len(arg.Datfile) == 0 {
        arg.Datfile = dat.FilePath()
    }
    
    if arg.Rng == 0 {
        arg.Rng = dat.Rng()
    }

    if arg.Azi == 0 {
        arg.Azi = dat.Azi()
    }
    
    arg.Type = dat.DataType()
        
    if arg.Flip {
        arg.LR = -1
    } else {
        arg.LR = 1
    }
    
    if arg.Min == 0.0 {
        arg.Min = 0.1
    }
    
    if arg.Max == 0.0 {
        arg.Min = 0.9
    }
    
    if arg.StartCC == 0 {
        arg.StartCC = 1
    }
    
    if arg.StartSec == 0 {
        arg.StartSec = 1
    }
    
    
    if arg.Cycle == 0 {
        arg.Cycle = 160.0
    }
    
    if arg.Elev == 0.0 {
        arg.Elev = 45.0
    }
    
    if arg.Orient == 0.0 {
        arg.Orient = 135.0
    }
    
    // TODO: implement proper deduction of plot mode
    //if arg.Mode == Undefined {
        //switch opt.DType {
            
        //}
    //}
    
    if arg.Inverse == 0 {
        arg.Inverse = Float2Raster 
    }
    
    if arg.Channel == 0 {
        arg.Channel = Red
    }
    
    if arg.CCMin == 0.0 {
        arg.CCMin = 0.2
    }
    
    //if op.colPost == 0.0 {
        //op.colPost, err = d.Float()
    //}
}
