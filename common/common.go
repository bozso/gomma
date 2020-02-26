package common

import (
    "log"
    "os"
    "math"
    "path"
    "strings"
    "encoding/json"
    "path/filepath"
    "fmt"
    
    "github.com/bozso/gamma/utils"
    "github.com/bozso/gamma/utils/io"
)

const DefaultCachePath = "/mnt/bozso_i/cache"

type RngAzi struct {
    Rng int `json:"rng"`
    Azi int `json:"azi"`
}

var DefaultRngAzi = RngAzi{Rng:1, Azi:1}

func (ra RngAzi) String() string {
    return fmt.Sprintf("%d,%d", ra.Rng, ra.Azi)
}

func (ra *RngAzi) Set(s string) (err error) {
    if len(s) == 0 {
        return utils.EmptyStringError{}
    }
    
    split, err := utils.NewSplitParser(s, ",")
    if err != nil { return }
    
    ra.Rng, err = split.Int(0)
    if err != nil { return }

    ra.Azi, err = split.Int(1)
    
    return
}

func (ra *RngAzi) Default() {
    if ra.Rng == 0 {
        ra.Rng = 1
    }
    
    if ra.Azi == 0 {
        ra.Azi = 1
    }
}

func (ra RngAzi) Check() (err error) {
    if ra.Rng == 0 {
        return ZeroDimError{dim: "range samples / columns"}
    }
    
    if ra.Azi == 0 {
        return ZeroDimError{dim: "azimuth lines / rows"}
    }
    
    return nil
}

type Dims interface {
    Rng() int
    Azi() int
}

type ZeroDimError struct {
    dim string
    err error
}

func (e ZeroDimError) Error() string {
    return fmt.Sprintf("expected %s to be non zero", e.dim)
}

func (e ZeroDimError) Unwrap() error {
    return e.err
}


type (
    Minmax struct {
        Min float64 `name:"min" default:"0.0"`
        Max float64 `name:"max" default:"1.0"`
    }
    
    IMinmax struct {
        Min int `name:"min" default:"0"`
        Max int `name:"max" default:"1"`
    }
    
    LatLon struct {
        Lat float64 `name:"lan" default:"1.0"`
        Lon float64 `name:"lot" default:"1.0"`
    }

    Slice []string
    GammaFun map[string]utils.CmdFun

    settings struct {
        RasExt    string
        Path      string
        Modules   []string
    }

    Point struct {
        X, Y float64
    }
    
    AOI [4]Point
    
    Rectangle struct {
        Max, Min Point
    }
)

const (
    BufSize    = 50
)

var (
    Pols = [4]string{"vv", "hh", "hv", "vh"}
    
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

    result := make(map[string]utils.CmdFun)

    for _, module := range Settings.Modules {
        for _, dir := range [2]string{"bin", "scripts"} {

            _path := filepath.Join(Path, module, dir, "*")
            glob, err := filepath.Glob(_path)

            if err != nil {
                utils.Fatal(err, "Glob '%s' failed! %s", _path, err)
            }

            for _, path := range glob {
                result[filepath.Base(path)] = utils.MakeCmd(path)
            }
        }
    }

    return result
}

func (self GammaFun) SelectFun(name1, name2 string) utils.CmdFun {
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

func (self GammaFun) Must(name string) (ret utils.CmdFun) {
    ret, ok := self[name]
    
    if !ok {
        log.Fatalf("failed to find Gamma executable '%s'", name)
    }
    
    return
}


func NoExt(p string) string {
    return strings.TrimSuffix(p, path.Ext(p))
}


func (p Point) InRect(r Rectangle) bool {
    return (p.X < r.Max.X && p.X > r.Min.X &&
            p.Y < r.Max.Y && p.Y > r.Min.Y)
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

func (mm *IMinmax) Set(s string) (err error) {
    if len(s) == 0 {
        return utils.EmptyStringError{}
    }
    
    split, err := utils.NewSplitParser(s, ",")
    if err != nil {
        return
    }
    
    mm.Min, err = split.Int(0)
    if err != nil {
        return
    }
    
    mm.Max, err = split.Int(1)
    if err != nil {
        return
    }
    
    return nil
}

func (ll LatLon) String() string {
    return fmt.Sprintf("%f,%f", ll.Lon, ll.Lat)
}

func (ll *LatLon) Set(s string) (err error) {
    var ferr = merr.Make("LatLon.Decode")

    if len(s) == 0 {
        return ferr.Wrap(utils.EmptyStringError{})
    }
    
    split, err := utils.NewSplitParser(s, ",")
    if err != nil {
        return
    }
    
    ll.Lat, err = split.Float(0)
    if err != nil {
        return
    }

    ll.Lon, err = split.Float(1)
    if err != nil {
        return
    }
    
    return nil
}

func SaveJson(path string, val interface{}) (err error) {
    out, err := json.MarshalIndent(val, "", "    ")
    if err != nil {
        return utils.WrapFmt(err, "failed to json encode struct: %v", val)
    }

    f, err := os.Create(path)
    if err != nil {
        return io.CreateFail(path, err)
    }
    defer f.Close()

    if _, err = f.Write(out); err != nil {
        return io.WriteFail(path, err)
    }

    return nil
}

func LoadJson(path string, val interface{}) (err error) {
    d, err := io.ReadFile(path)
    if err != nil {
        return utils.WrapFmt(err, "failed to read file '%s'", path)
    }
    
    if err := json.Unmarshal(d, &val); err != nil {
        return utils.WrapFmt(err, "failed to parse json data %s'", d)
    }

    return nil
}
