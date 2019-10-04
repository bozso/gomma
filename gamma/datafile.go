package gamma

import (
    "fmt"
    "os"

    //"time"
    fp "path/filepath"
)

type (
    dataFile struct {
        Dat   string
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
    return Params{Par: path, sep: ":", contents: nil}
}

func NewDataFile(dat, par, ext string) (ret dataFile, err error) {
    if len(dat) == 0 {
        err = Handle(err, "'dat' should not be an empty string")
        return
    }
    
    ret.Dat = dat
    
    if len(ext) == 0 {
        ext = "par"
    }
    
    if len(par) == 0 {
        par = fmt.Sprintf("%s.%s", dat, ext)
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
    return d.Dat
}

func (d dataFile) Parfile() string {
    return d.Par
}

func (d dataFile) Rng() (int, error) {
    return d.Int("range_samples")
}

func (d dataFile) Azi() (int, error) {
    return d.Int("azimuth_lines")
}

func (d dataFile) ImageFormat() (string, error) {
    return d.Param("image_format")
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

func (opt *MLIOpt) Parse() {
    opt.ScaleExp.Parse()
    
    if len(opt.refTab) == 0 {
        opt.refTab = "-"
    }
    
    if opt.Looks.Rng == 0 {
        opt.Looks.Rng = 1
    }
    
    if opt.Looks.Azi == 0 {
        opt.Looks.Azi = 1
    }
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


// TODO: implement proper selection of plot command
func Raster(dat DataFile, opt rasArgs, sec string) (err error) {
    err = opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse display options")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("ras" + cmd)
    
    switch cmd {
        case "SLC":
            err = rasslc(opt)
            
            if err != nil {
                return
            }
            
        case "MLI":
            err = raspwr(opt)
            
            if err != nil {
                return
            }
        
        default:
            err = Handle(nil, "unrecognized command type '%s'", cmd)
            return
    }
    
    
    
    
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
