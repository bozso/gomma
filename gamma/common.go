package gamma

import (
    "log"
    "math"
	//"fmt"
	//"os"
	fp "path/filepath"
    pt "path"
    str "strings"
)

type (
    GammaFun map[string]CmdFun

	settings struct {
		RasExt    string
		path      string
		modules   []string
	}

	Point struct {
		X, Y float64
	}
    
    AOI [4]Point
    
	Rect struct {
		Max, Min Point
	}

	disArgs struct {
		flip, debug        bool
		rng, azi           int
		imgFormat, datfile string
	}

	rasArgs struct {
		disArgs
		ext     string
		avgFact int
	}
)

const (
	useVersion = "20181130"
	BufSize    = 50
)

var (
	versions = map[string]string{
		"20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
	}

	Pols = [4]string{"vv", "hh", "hv", "vh"}

	DataTypes = map[string]int{
		"FCOMPLEX":  0,
		"SCOMPLEX":  1,
		"FLOAT":     0,
		"SHORT_INT": 1,
		"DOUBLE":    2,
	}
	Gamma = makeGamma()
	Imv   = MakeCmd("eog")

	Settings = settings{
		RasExt:  "bmp",
		path:    versions[useVersion],
		modules: []string{"DIFF", "DISP", "ISP", "LAT", "IPTA"},
	}
)

func makeGamma() GammaFun {
	Path := Settings.path

	result := make(map[string]CmdFun)

	for _, module := range Settings.modules {
		for _, dir := range [2]string{"bin", "scripts"} {

			_path := fp.Join(Path, module, dir, "*")
			glob, err := fp.Glob(_path)

			if err != nil {
				Fatal(err, "makeGamma: Glob '%s' failed! %w", _path, err)
			}

			for _, path := range glob {
				result[fp.Base(path)] = MakeCmd(path)
			}
		}
	}

	return result
}

func (self GammaFun) selectFun(name1, name2 string) CmdFun {
    ret, ok := self[name1]
    
    if ok {
        return ret
    }
    
    ret, ok = self[name2]
    
    if !ok {
        log.Fatalf("Either '%s' ot '%s' must be an available executable!",
            name1, name2)
        return nil
    }
    return ret
}

func (self GammaFun) must(name string) (ret CmdFun) {
    ret, ok := self[name]
    
    if !ok {
        log.Fatalf("Gamma executable '%s' does not exist!", name)
    }
    return
}


func NoExt(path string) string {
    return str.TrimSuffix(path, pt.Ext(path))
}


func (self *Point) InRect(r *Rect) bool {
	return (self.X < r.Max.X && self.X > r.Min.X &&
            self.Y < r.Max.Y && self.Y > r.Min.Y)
}

func First() string {
	return "First"
}

func ParseDisArgs(d dataFile, args disArgs) (ret *disArgs, err error) {
	handle := Handler("ParseDisArgs")

	if len(args.datfile) == 0 {
		args.datfile = d.dat
	}

	if args.rng == 0 {
		if args.rng, err = d.Rng(); err != nil {
			err = handle(err, "Could not get range_samples!")
            return
		}
	}

	if args.azi == 0 {
		if args.azi, err = d.Azi(); err != nil {
			err = handle(err, "Could not get azimuth_lines!")
            return
		}
	}

	// parts = pth.basename(datfile).split(".")
	if len(args.imgFormat) == 0 {
		if args.imgFormat, err = d.imgFormat(); err != nil {
			err = handle(err, "Could not get image_format!")
            return
		}
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

	return &args, nil
}

func isclose(num1, num2 float64) bool {
    return math.RoundToEven(math.Abs(num1 - num2)) > 0.0
}
