package gamma

import (
    "fmt"
    "os"

    //"time"
    fp "path/filepath"
)

type (
    dataFile struct {
        dat   string
        files []string
        Params
        date
    }

    DataFile interface {
        Datfile() string
        Parfile() string
        Rng() (int, error)
        Azi() (int, error)
        Int(string) (int, error)
        Float(string) (float64, error)
        PlotCmd() string
        ImageFormat() (string, error)
        //Display(disArgs) error
        //Raster(rasArgs) error
    }

    SLC struct {
        dataFile
    }

    MLI struct {
        dataFile
    }
    
    // TODO: add loff, nlines
    MLIOpt struct {
        refTab string
        Looks RngAzi
        windowFlag bool
        ScaleExp
    }
)

func NewGammaParam(path string) Params {
    return Params{par: path, sep: ":", contents: nil}
}

func NewDataFile(dat, par, ext string) (ret dataFile, err error) {
    if len(dat) == 0 {
        err = Handle(err, "'dat' should not be an empty string")
        return
    }
    
    ret.dat = dat
    
    exist, err := Exist(dat)
    
    if err != nil {
        err = Handle(err, "failed to check whether datafile '%s' exists",
            dat)
        return
    }
    
    if !exist {
        err = Handle(nil, "datafile '%s' does not exist", dat)
        return
    }
    
    if len(par) == 0 {
        par = "par"
    }
    
    if len(par) == 0 {
        par = fmt.Sprintf("%s.%s", dat, ext)
    }
    
    exist, err = Exist(par)
    
    if err != nil {
        err = Handle(err, "failed to check whether parfile '%s' exists",
            dat)
        return
    }
    
    if !exist {
        err = Handle(nil, "parfile '%s' does not exist", dat)
        return
    }
    
    ret.Params = NewGammaParam(par)
    ret.files = []string{dat, par}

    return ret, nil
}

func NewSLC(dat, par string) (ret SLC, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

func NewMLI(dat, par string) (ret MLI, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

func (d *dataFile) Exist() (ret bool, err error) {
    var exist bool
    for _, file := range d.files {
        exist, err = Exist(file)

        if err != nil {
            err = Handle(err, "stat on file '%s' failed", file)
            return
        }

        if !exist {
            return false, nil
        }
    }
    return true, nil
}

func (d *dataFile) Move(path string) error {
    for _, file := range d.files {
        if len(file) == 0 {
            continue
        }

        dst := fp.Join(path, file)
        err := os.Rename(file, dst)

        if err != nil {
            return Handle(err, "failed to move file '%s' to '%s'", file, dst)
        }
    }
    return nil
}

func (d dataFile) Datfile() string {
    return d.dat
}

func (d dataFile) Parfile() string {
    return d.par
}

func (d dataFile) Rng() (int, error) {
    return d.Int("range_samples")
}

func (d dataFile) Azi() (int, error) {
    return d.Int("azimuth_lines")
}

func (d dataFile) ImageFormat() (string, error) {
    return d.Par("image_format")
}

func (d dataFile) PlotCmd() string {
    return ""
}

func (d SLC) PlotCmd() string {
    return "SLC"
}

func (d MLI) PlotCmd() string {
    return "MLI"
}

func (arg *disArgs) Parse(dat DataFile) (err error) {
    if len(arg.Datfile) == 0 {
        arg.Datfile = dat.Datfile()
    }
    
    if len(arg.Cmd) == 0 {
        arg.Cmd = dat.PlotCmd()
    }

    if arg.Rng == 0 {
        if arg.Rng, err = dat.Rng(); err != nil {
            return Handle(err, "failed to get range_samples")
        }
    }

    if arg.Azi == 0 {
        if arg.Azi, err = dat.Azi(); err != nil {
            return Handle(err, "failed to get azimuth_lines")
        }
    }

    // parts = pth.basename(datfile).split(".")
    if len(arg.ImgFmt) == 0 {
        if arg.ImgFmt, err = dat.ImageFormat(); err != nil {
            return Handle(err, "failed to get image_format")
        }
    }

    if arg.Flip {
        arg.LR = 1
    } else {
        arg.LR = 0
    }

    return nil
}

// TODO: Finish
func (opt *rasArgs) Parse(dat DataFile) error {
    err := opt.disArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display arguments")
    }
    
    if opt.avgFact == 0 {
        opt.avgFact = 1000
    }
    
    rng := opt.Avg.Rng
    
    if rng == 0 {
        rng = opt.Rng / opt.avgFact
        
        if rng < 0 {
            rng = 1
        }
    }
    
    opt.Avg.Rng = rng
    
    azi := opt.Avg.Azi
    
    if azi == 0 {
        azi = opt.Azi / opt.avgFact
        
        if azi < 0 {
            azi = 1
        }
    }
    
    opt.Avg.Azi = azi
    
    if len(opt.raster) == 0 {
        opt.raster = fmt.Sprintf("%s.%s", opt.Datfile, Settings.RasExt)
    }
    
    return nil
}

func (s *SLC) Raster(opt rasArgs) error {
    err := opt.Parse(s)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return rasslc(opt)
}

func (m *MLI) Raster(opt rasArgs) error {
    err := opt.Parse(m)
    
    if err != nil {
        return Handle(err, "failed to parse raster options")
    }
    
    return raspwr(opt)
}

func Display(dat DataFile, opt disArgs) error {
    err := opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("dis" + cmd)
    
    if cmd == "SLC" {
        _, err := fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines, opt.Scale,
                      opt.Exp)
        
        if err != nil {
            return Handle(err, "failed to execute display command")
        }
    }
    return nil
}

func Raster(dat DataFile, opt rasArgs, sec string) (err error) {
    err = opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("ras" + cmd)

    if cmd == "SLC" {
        _, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
            opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
            opt.ImgFmt, opt.headerSize, opt.raster)

    } else {
        if len(sec) == 0 {
            _, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.raster, opt.ImgFmt, opt.headerSize)

        } else {
            _, err = fun(opt.Datfile, sec, opt.Rng, opt.Start, opt.Nlines,
                opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
                opt.LR, opt.raster, opt.ImgFmt, opt.headerSize, opt.raster)
        }
    }
    
    if err != nil {
        return Handle(err, "failed to create rasterfile '%s'", opt.raster)
    }
    
    return nil
}
