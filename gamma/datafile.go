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
		Int() (int, error)
		Float() (float64, error)
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

	disArgs struct {
		Flip                 bool
		ImgFmt, Datfile, Cmd string
		Start, Nlines, LR    int
		Scale, Exp           float64
		RngAzi
	}

	rasArgs struct {
		disArgs
		//ext                 string
		avgFact, headerSize int
		Avg                 RngAzi
	}
)

func NewGammaParam(path string) Params {
	return Params{par: path, sep: ":", contents: nil}
}

func NewDataFile(dat, par string) (ret dataFile, err error) {
	ret.dat = dat

	if len(dat) == 0 {
		err = Handle(err, "'dat' should not be an empty string: '%s'", dat)
        return
	}

	if len(par) == 0 {
		par = dat + ".par"
	}

	ret.Params = NewGammaParam(par)

	ret.files = []string{dat, par}

	return ret, nil
}

func NewSLC(dat, par string) (ret SLC, err error) {
	ret.dataFile, err = NewDataFile(dat, par)
	return
}

func (d *dataFile) Exist() (ret bool, err error) {
	var exist bool
	for _, file := range d.files {
		exist, err = Exist(file)

		if err != nil {
			err = fmt.Errorf("Stat on file '%s' failed!\nError: %w!",
				file, err)
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
			return Handle(err, "Failed to move file '%s' to '%s'!", file, dst)
		}
	}
	return nil
}

func (d *dataFile) Datfile() string {
	return d.dat
}

func (d *dataFile) Parfile() string {
	return d.par
}

func (d *dataFile) Rng() (int, error) {
	return d.Int("range_samples")
}

func (d *dataFile) Azi() (int, error) {
	return d.Int("azimuth_samples")
}

func (d *dataFile) ImageFormat() (string, error) {
	return d.Par("image_format")
}

func (d *SLC) PlotFun() string {
	return "SLC"
}

func (arg *disArgs) Parse(dat DataFile) (err error) {
	if len(arg.Datfile) == 0 {
		arg.Datfile = dat.Datfile()
	}
	
    arg.Cmd = dat.PlotCmd()

	if arg.Rng == 0 {
		if arg.Rng, err = dat.Rng(); err != nil {
			return Handle(err, "Could not get range_samples!")
		}
	}

	if arg.Azi == 0 {
		if arg.Azi, err = dat.Azi(); err != nil {
			return Handle(err, "Could not get azimuth_lines!")
		}
	}

	// parts = pth.basename(datfile).split(".")
	if len(arg.ImgFmt) == 0 {
		if arg.ImgFmt, err = dat.ImageFormat(); err != nil {
			return Handle(err, "Could not get image_format!")
		}
	}

	if arg.Flip {
		arg.LR = 1
	} else {
		arg.LR = 0
	}

	// args.flip = -1 if flip else 1

	/*
	   if cmd is None:
	       try:
	           ext = [ext for ext in parts if ext in extensions][0]
	       except IndexError:
	           raise ValueError("Unrecognized extension of file %s. Available "
	                            "extensions: %s" % (datfile, pr.extensions))
	       cmd = [cmd for cmd, exts in plot_cmd_files.items()
	              if ext in exts][0]
	*/

	return nil
}

// TODO: Finish
func (opt *rasArgs) Parse(dat DataFile) error {
	err := opt.disArgs.Parse(dat)

	if err != nil {
		return Handle(err, "Failed to parse display arguments!")
	}

	return nil
}

func Display(dat DataFile, opt disArgs) error {
	err := opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "Failed to parse display options!")
    }
    
    cmd := opt.Cmd
    fun := Gamma.must("dis" + cmd)
    
	if cmd == "SLC" {
		_, err := fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines, opt.Scale,
                      opt.Exp)
        
        if err != nil {
            return Handle(err, "Failed to execute display command!")
        }
	}
    return nil
}

func Raster(dat DataFile, opt rasArgs, sec string) (err error) {
	err = opt.Parse(dat)
    
    if err != nil {
        return Handle(err, "Failed to parse display options!")
    }
	
    cmd := opt.Cmd
    fun := Gamma.must("dis" + cmd)

	raster := fmt.Sprintf("%s.%s", dat.Datfile(), Settings.RasExt)

	if cmd == "SLC" {
		_, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
			opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
			opt.ImgFmt, opt.headerSize, raster)

	} else {
		if len(sec) == 0 {
			_, err = fun(opt.Datfile, opt.Rng, opt.Start, opt.Nlines,
				opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
				opt.LR, raster, opt.ImgFmt, opt.headerSize)

		} else {
			_, err = fun(opt.Datfile, sec, opt.Rng, opt.Start, opt.Nlines,
				opt.Avg.Rng, opt.Avg.Azi, opt.Scale, opt.Exp,
				opt.LR, raster, opt.ImgFmt, opt.headerSize)
		}
	}
    
    if err != nil {
        return Handle(err, "Failed to create rasterfile '%s'!", raster)
    }
    
    return nil
}
