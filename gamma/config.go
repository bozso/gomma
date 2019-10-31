package gamma

import (
    "encoding/json"
    "fmt"
    "os"
    "log"
    //ref "reflect"
    //conv "strconv"
    str "strings"
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
    
    GeneralOpt struct {
        DataPath, OutputDir, Pol string
        MasterDate               string
        CachePath                string `json:"CACHE_PATH"`
        Looks                    RngAzi
        IWs                      [3]IMinmax
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

    CoregOpt struct {
        CoherenceThresh  float64 `name:"coh"    default:"0.8"`
        FractionThresh   float64 `name:"frac"   default:"0.01"`
        PhaseStdevThresh float64 `name:"phase"  default:"0.8"`
        MasterIdx        int     `name:"master" default:"0"`
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
        infile        string
        General       GeneralOpt
        PreSelect     PreSelectOpt
        Geocoding     GeocodeOpt
        Coreg         CoregOpt
        IFGSelect     IfgSelectOpt
        CalcCoherence CoherenceOpt
    }

    stepFun func(*Config) error
)

const (
    DefaultCachePath = "/mnt/bozso_i/cache"
)

var (
    steps = map[string]stepFun{
        "select": stepSelect,
        "import": stepImport,
        "geo": stepGeocode,
        //"coreg":  stepCoreg,
    }

    stepList = MapKeys(steps)

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

        Coreg: CoregOpt{
            CoherenceThresh:  0.8,
            FractionThresh:   0.01,
            PhaseStdevThresh: 0.8,
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

func (ra *RngAzi) Default() {
    if ra.Rng == 0 {
        ra.Rng = 1
    }
    
    if ra.Azi == 0 {
        ra.Azi = 1
    }
}

func delim(msg, sym string) {
    msg = fmt.Sprintf("%s %s %s", sym, msg, sym)
    syms := str.Repeat(sym, len(msg))

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

func SaveJson(path string, val interface{}) error {
    out, err := json.MarshalIndent(val, "", "    ")
    if err != nil {
        return Handle(err, "failed to json encode struct: %v", val)
    }

    f, err := os.Create(path)
    if err != nil {
        return Handle(err, "failed to create file: %s", path)
    }
    defer f.Close()

    if _, err = f.Write(out); err != nil {
        return Handle(err, "failed to write to file '%s'", path)
    }

    return nil
}

func LoadJson(path string, val interface{}) error {
    data, err := ReadFile(path)

    if err != nil {
        return Handle(err, "failed to read file '%s'", path)
    }
    
    if err := json.Unmarshal(data, &val); err != nil {
        return Handle(err, "failed to parse json data %s'", data)
    }

    return nil
}
