package gamma

import (
    "log"
    "math"
    "path"
    "strings"
    "path/filepath"
    //"os"
)

var merr = NewModuleErr("gamma")

type (
    Slice []string
    GammaFun map[string]CmdFun

    settings struct {
        RasExt    string
        Path      string
        Modules   []string
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
    // TODO: deprecate
    //versions = map[string]string{
    //    "20181130": "/home/istvan/progs/GAMMA_SOFTWARE-20181130",
    //}

    Pols = [4]string{"vv", "hh", "hv", "vh"}
    
    // TODO: deprecate this
    DataTypes = map[string]int{
        "FCOMPLEX":  0,
        "SCOMPLEX":  1,
        "FLOAT":     0,
        "SHORT_INT": 1,
        "DOUBLE":    2,
    }
    
    // TODO: get settings path from environment variable
    Settings = loadSettings("/home/istvan/progs/gamma/bin/settings.json")
    Gamma = makeGamma()
)

func loadSettings(path string) (ret settings) {
    if err := LoadJson(path, &ret); err != nil {
        log.Fatalf("Failed to load Gamma settings from '%s'\nError:'%s'\n!",
            path, err)
    }
    
    return
}

func makeGamma() GammaFun {
    Path := Settings.Path

    result := make(map[string]CmdFun)

    for _, module := range Settings.Modules {
        for _, dir := range [2]string{"bin", "scripts"} {

            _path := filepath.Join(Path, module, dir, "*")
            glob, err := filepath.Glob(_path)

            if err != nil {
                Fatal(err, "Glob '%s' failed! %s", _path, err)
            }

            for _, path := range glob {
                result[filepath.Base(path)] = MakeCmd(path)
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
        log.Fatalf("either '%s' or '%s' must be an available executable",
            name1, name2)
    }
    
    return ret
}

func (self GammaFun) Must(name string) (ret CmdFun) {
    ret, ok := self[name]
    
    if !ok {
        log.Fatalf("failed to find Gamma executable '%s'", name)
    }
    
    return
}


func NoExt(p string) string {
    return strings.TrimSuffix(p, path.Ext(p))
}


func (self *Point) InRect(r *Rect) bool {
    return (self.X < r.Max.X && self.X > r.Min.X &&
            self.Y < r.Max.Y && self.Y > r.Min.Y)
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
