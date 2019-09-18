package gamma

import (
	"encoding/json"
	"fmt"
	"os"
    "log"
	ref "reflect"
	//conv "strconv"
	str "strings"
)

type (
	minmax struct {
		Min, Max float64
	}
    
    LatLon struct {
        Lat, Lon float64
    }
    
    RngAzi struct {
        Rng, Azi int
    }
    
	general struct {
		DataPath, OutputDir, Pol, Metafile string
        CachePath                          string `json:"CACHE_PATH"`
		Looks                              RngAzi
	}

	preselect struct {
		DateStart, DateStop   string
        LowerLeft, UpperRight LatLon
		CheckZips             bool
	}

	geocoding struct {
		DEMPath      string
		Iter, NPoly  int
        DEMOverlap      RngAzi
        DEMOverSampling LatLon
	}

	coreg struct {
		CoherenceThresh, FractionThresh, PhaseStdevThresh float64
	}

	ifgSelect struct {
		Bperp  minmax
		DeltaT minmax
	}

	coherence struct {
		WeightType             string
		Box                    minmax
		SlopeCorrelationThresh float64
		SlopeWindow            int
	}

	config struct {
		General   general
		PreSelect preselect
		Geocoding geocoding
		Coreg     coreg
		IFGSelect ifgSelect
		Coherence coherence
	}

	stepFun func(*config) error
)

const (
    DefaultCachePath = "/mnt/bozso_i/cache"
)

var (
	steps = map[string]stepFun{
		"select": stepSelect,
        //"coreg":  stepCoreg,
	}

	stepList []string

	defaultConfig = config{
		General: general{
			Pol: "vv",
            Metafile: "meta.json",
            OutputDir: ".",
			Looks: RngAzi{
                Rng: 1,
                Azi: 1,
            },
		},

		PreSelect: preselect{
			CheckZips:  false,
		},

		Geocoding: geocoding{
			DEMPath:            "/home/istvan/DEM/srtm.vrt",
			Iter: 1,
			DEMOverlap: RngAzi{
                Rng: 100,
                Azi: 100,
            },
			NPoly: 1,
			DEMOverSampling: LatLon{
                Lat: 2.0,
                Lon: 2.0,
            },
		},

		Coreg: coreg{
			CoherenceThresh:  0.8,
			FractionThresh:   0.01,
			PhaseStdevThresh: 0.8,
		},

		IFGSelect: ifgSelect{
			Bperp:  minmax{Min: 0.0, Max: 150.0},
			DeltaT: minmax{Min: 0.0, Max: 15.0},
		},

		Coherence: coherence{
			WeightType:             "gaussian",
			Box:                    minmax{Min: 3.0, Max: 9.0},
			SlopeCorrelationThresh: 0.4,
			SlopeWindow:            5,
		},
	}
)


func init() {
	keys := ref.ValueOf(steps).MapKeys()

	stepList = make([]string, len(keys))

	for ii, key := range keys {
		stepList[ii] = key.String()
	}
}

func delim(msg, sym string) {
	msg = fmt.Sprintf("%s %s %s", sym, msg, sym)
	syms := str.Repeat(sym, len(msg))

	log.Printf("\n%s\n%s\n%s\n", syms, msg, syms)
}


func MakeDefaultConfig(path string) error {
	handle := Handler("MakeDefaultConfig")

	out, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		return handle(err, "Failed to json encode default configuration!")
	}

	f, err := os.Create(path)
	if err != nil {
		return handle(err, "Failed to create file: %v!", path)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		return handle(err, "Failed to write to file '%v'!", path)
	}

	return nil
}

func SaveJson(path string, val interface{}) error {
	handle := Handler("SaveJson")
	
    out, err := json.MarshalIndent(val, "", "    ")
	if err != nil {
		return handle(err, "Failed to json encode struct!")
	}

	f, err := os.Create(path)
	if err != nil {
		return handle(err, "Failed to create file: %v!", path)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		return handle(err, "Failed to write to file '%v'!", path)
	}

	return nil
    
}

func LoadJson(path string, val interface{}) error {
	handle := Handler("LoadJson")
	
	data, err := ReadFile(path)

	if err != nil {
		return handle(err, "Failed to read file:  '%s'!", path)
	}
    
	if err := json.Unmarshal(data, &val); err != nil {
		return handle(err, "Failed to parse json data: %s'!", data)
	}

	return nil
}
