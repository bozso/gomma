package gamma

import (
	"encoding/json"
	"fmt"
	"os"
	fl "flag"
	ref "reflect"
	//conv "strconv"
	str "strings"
)

type (
	minmax struct {
		Min, Max float64
	}
    
    latLon struct {
        Lat, Lon float64
    }
    
	general struct {
		CachePath                          string `json:"CACHE_PATH,omitempty"`
		DataPath, OutputDir, Pol, Metafile string
		RangeLooks, AzimuthLooks           int
	}

	preselect struct {
		DateStart, DateStop, MasterDate string
        LowerLeft, UpperRight           latLon
		CheckZips                       bool
	}

	geocoding struct {
		DEMPath                                   string
		Iter, RangeOverlap, AzimuthOverlap, NPoly int
		DEMLatOversampling, DEMLonOversampling    float64
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

	cliConfig struct {
		Conf, Step, Start, Stop, Log string
		Skip, Show                   bool
	}
    
    S1ProcData struct {
        MasterDate string
        Zipfiles   S1Zips
    }
    
	stepFun func(*config) error
)


var (
	steps = map[string]stepFun{
		"preselect": stepPreselect,
		"coreg":     stepCoreg,
	}

	stepList []string

	defaultConfig = config{
		General: general{
			CachePath:    "/mnt/bozso_i/cache",
			Pol:          "vv",
			RangeLooks:   1,
			AzimuthLooks: 1,
		},

		PreSelect: preselect{
			MasterDate: "auto",
			CheckZips:  false,
		},

		Geocoding: geocoding{
			DEMPath:            "/home/istvan/DEM/srtm.vrt",
			Iter:               1,
			RangeOverlap:       100,
			AzimuthOverlap:     100,
			NPoly:              1,
			DEMLatOversampling: 2.0,
			DEMLonOversampling: 2.0,
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

func center(s string, n int, fill string) string {
	div := n / 2
	return str.Repeat(fill, div) + s + str.Repeat(fill, div)
}

const width = 40

func delim(msg, sym string) {
	msg = fmt.Sprintf("%s %s %s", sym, msg, sym)
	syms := center(str.Repeat(sym, len(msg)), width, " ")
	msg = center(msg, width, " ")

	fmt.Printf("%s\n%s\n%s\n", syms, msg, syms)
}

func NewConfig(flag *fl.FlagSet) *cliConfig {
	conf := cliConfig{}

	flag.StringVar(&conf.Conf, "config", "gamma.json",
		"Processing configuration file")

	flag.StringVar(&conf.Step, "step", "",
		"Single processing step to be executed.")

	flag.StringVar(&conf.Start, "start", "",
		"Starting processing step.")

	flag.StringVar(&conf.Stop, "stop", "",
		"Last processing step to be executed.")

	flag.StringVar(&conf.Log, "logfile", "gamma.log",
		"Log messages will be saved here.")

	flag.BoolVar(&conf.Skip, "skip_optional", false,
		"If set the proccessing will skip optional steps.")
	flag.BoolVar(&conf.Show, "show_steps", false,
		"If set, prints the processing steps.")

	return &conf
}

func stepIndex(step string) int {
	for ii, _step := range stepList {
		if step == _step {
			return ii
		}
	}
	return -1
}

func listSteps() {
	fmt.Println("Available processing steps: ", stepList)
}

func (self *cliConfig) Parse() (config, int, int, error) {
	handle := Handler("CLIConfig.Parse")

	if self.Show {
		listSteps()
		os.Exit(0)
	}

	istep, istart, istop := stepIndex(self.Step), stepIndex(self.Start),
		stepIndex(self.Stop)

	if istep == -1 {
		if istart == -1 {
			listSteps()
			return config{}, 0, 0,
				handle(nil,
					"Starting step '%s' is not in list of available steps!",
					self.Start)

		}

		if istop == -1 {
			listSteps()
			return config{}, 0, 0,
				handle(nil,
					"Stopping step '%s' is not in list of available steps!",
					self.Stop)

		}
	} else {
		istart = istep
		istop = istep + 1
	}

	path := self.Conf

	var ret config

	data, err := ReadFile(path)

	if err != nil {
		return config{}, 0, 0, handle(err, "Failed to read file:  '%s'!",
			path)
	}

	if err := json.Unmarshal(data, &ret); err != nil {
		return config{}, 0, 0, handle(err, "Failed to parse json data: %s'!",
			data)
	}

	return ret, istart, istop, nil
}

func (self *config) RunSteps(start, stop int) error {
	handle := Handler("RunSteps")

	for ii := start; ii < stop; ii++ {
		name := stepList[ii]
		step, _ := steps[name]

		delim(fmt.Sprintf("START: %s", name), "*")

		if err := step(self); err != nil {
			return handle(err, "Error while running step: '%s'",
				name)
		}

		delim(fmt.Sprintf("END: %s", name), "*")
	}
	return nil
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
