package gamma

import (
    "encoding/json"
    "fmt"
    "os"
    "log"
    "strings"
)

type (
    /*
    PreSelectOpt struct {
        DateStart, DateStop   string
        LowerLeft, UpperRight LatLon
        CheckZips             bool
    }
    
    GeocodeOpt struct {
        DEMPath                   string
        Iter, NPoly, nPixel       int
        LanczosOrder, MLIOversamp int
        DEMOverlap, OffsetWindows common.RngAzi
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
    */
)

const (
    DefaultCachePath = "/mnt/bozso_i/cache"
)

var (
    /*
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
    */
)



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
