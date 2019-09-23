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
    Slice []string
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
				Fatal(err, "makeGamma: Glob '%s' failed! %s", _path, err)
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
        log.Fatalf("Either '%s' or '%s' must be an available executable!",
            name1, name2)
    }
    
    return ret
}

func (self GammaFun) must(name string) (ret CmdFun) {
    ret, ok := self[name]
    
    if !ok {
        log.Fatalf("Could not find Gamma executable '%s'!", name)
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

func isclose(num1, num2 float64) bool {
    return math.RoundToEven(math.Abs(num1 - num2)) > 0.0
}

func (sl Slice) Contains(s string) bool {
    for _, elem := range sl {
        if s == elem {
            return true
        }
    }
    return false
}
