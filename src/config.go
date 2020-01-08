package gamma

import (
    "encoding/json"
    "fmt"
    "os"
    "log"
    "strings"
)

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
    
    PreSelectOpt struct {
        DateStart, DateStop   string
        LowerLeft, UpperRight LatLon
        CheckZips             bool
    }
    
    GeocodeOpt struct {
        DEMPath                   string
        Iter, NPoly, nPixel       int
        LanczosOrder, MLIOversamp int
        DEMOverlap, OffsetWindows RngAzi
        DEMOverSampling           LatLon
        AreaFactor, CCThresh      float64
        BandwithFrac, RngOversamp float64
        Master                    MLI
    }


    IfgSelectOpt struct {
        Bperp  Minmax
        DeltaT Minmax
    }

    CoherenceOpt struct {
        WeightType             string  `name:"weight" default:""`
        Box                    Minmax
        SlopeCorrelationThresh float64 `name:"slope"  default:""`
        SlopeWindow            int     `name:"win"    default:""`
    }

    Config struct {
        InFile        string
        General       GeneralOpt
        PreSelect     PreSelectOpt
        Geocoding     GeocodeOpt
        IFGSelect     IfgSelectOpt
        CalcCoherence CoherenceOpt
    }

    stepFun func(*Config) error
)

const (
    DefaultCachePath = "/mnt/bozso_i/cache"
)

var (
    defaultConfig = Config{
        General: GeneralOpt{
            Pol: "vv",
            OutputDir: ".",
            MasterDate: "",
            CachePath: "/mnt/storage_A/istvan/cache",
            Looks: RngAzi{
                Rng: 1,
                Azi: 1,
            },
        },

        PreSelect: PreSelectOpt{
            CheckZips:  false,
        },

        Geocoding: GeocodeOpt{
            DEMPath: "/mnt/storage_B/szucs_e/SRTMGL1/SRTM.vrt",
            Iter: 1,
            nPixel: 8,
            LanczosOrder: 5,
            NPoly: 1,
            MLIOversamp: 2,
            CCThresh: 0.1,
            BandwithFrac: 0.8,
            AreaFactor: 20.0,
            RngOversamp: 2.0,
            DEMOverlap: RngAzi{
                Rng: 100,
                Azi: 100,
            },
            DEMOverSampling: LatLon{
                Lat: 2.0,
                Lon: 2.0,
            },
            OffsetWindows: RngAzi{
                Rng: 500,
                Azi: 500,
            },
        },

        IFGSelect: IfgSelectOpt{
            Bperp:  Minmax{Min: 0.0, Max: 150.0},
            DeltaT: Minmax{Min: 0.0, Max: 15.0},
        },

        CalcCoherence: CoherenceOpt{
            WeightType:             "gaussian",
            Box:                    Minmax{Min: 3.0, Max: 9.0},
            SlopeCorrelationThresh: 0.4,
            SlopeWindow:            5,
        },
    }
)


func (mm *IMinmax) Set(s string) (err error) {
    var ferr = merr.Make("IMinmax.Decode")
    
    if len(s) == 0 {
        return ferr.Wrap(EmptyStringError{})
    }
    
    split := NewSplitParser(s, ",")
    
    mm.Min = split.Int(0)
    mm.Max = split.Int(1)
    
    if err = split.Wrap(); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

func (ll LatLon) String() string {
    return fmt.Sprintf("%f,%f", ll.Lon, ll.Lat)
}

func (ll *LatLon) Set(s string) (err error) {
    var ferr = merr.Make("LatLon.Decode")

    if len(s) == 0 {
        return ferr.Wrap(EmptyStringError{})
    }
    
    split := NewSplitParser(s, ",")
    
    ll.Lat = split.Float(0, 64)
    ll.Lon = split.Float(1, 64)
    
    if err = split.Wrap(); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

func delim(msg, sym string) {
    msg = fmt.Sprintf("%s %s %s", sym, msg, sym)
    syms := strings.Repeat(sym, len(msg))

    log.Printf("\n%s\n%s\n%s\n", syms, msg, syms)
}


func MakeDefaultConfig(path string) (err error) {
    out, err := json.MarshalIndent(defaultConfig, "", "    ")
    
    if err != nil {
        return Handle(err, "failed to json encode default configuration")
    }
        
    f, err := os.Create(path);
    if err != nil {
        return Handle(err, "failed to create file: %s", path)
    }
    defer f.Close()

    
    if _, err = f.Write(out); err != nil {
        return Handle(err, "failed to write to file '%s'", path)
    }

    return nil
}

func SaveJson(path string, val interface{}) (err error) {
    var ferr = merr.Make("SaveJson")
    
    var out []byte
    if out, err = json.MarshalIndent(val, "", "    "); err != nil {
        return ferr.WrapFmt(err,
            "failed to json encode struct: %v", val)
    }

    var f *os.File
    if f, err = os.Create(path); err != nil {
        return ferr.WrapFmt(err, "failed to create file: %s", path)
    }
    defer f.Close()

    if _, err = f.Write(out); err != nil {
        return ferr.WrapFmt(err, "failed to write to file '%s'", path)
    }

    return nil
}

func LoadJson(path string, val interface{}) (err error) {
    var ferr = merr.Make("LoadJson")
    
    var data []byte
    if data, err = ReadFile(path); err != nil {
        return ferr.WrapFmt(err, "failed to read file '%s'", path)
    }
    
    if err := json.Unmarshal(data, &val); err != nil {
        return ferr.WrapFmt(err, "failed to parse json data %s'", data)
    }

    return nil
}
