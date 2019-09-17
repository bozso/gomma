package gamma

import (
    //"log"
    "fmt"
    "os"
    "encoding/json"
    fp "path/filepath"
    fl "flag"
    //str "strings"
)


type(
    Process struct {
		Conf, Step, Start, Stop, Log, CachePath string
		Skip, Show                              bool
        config
    }
    
    Lister struct {
        conf, Mode string
        config
        Meta
    }
    
    Meta struct {
        MasterIdx  int
        MasterDate string
        Zipfiles   []string
    }
)    

var(
    ListModes = []string{"quicklook"}
)

func NewProcess(args []string) (ret Process, err error) {
    flag := fl.NewFlagSet("proc", fl.ContinueOnError)
    
	flag.StringVar(&ret.Conf, "config", "gamma.json",
		"Processing configuration file")

	flag.StringVar(&ret.Step, "step", "",
		"Single processing step to be executed.")

	flag.StringVar(&ret.Start, "start", "",
		"Starting processing step.")

	flag.StringVar(&ret.Stop, "stop", "",
		"Last processing step to be executed.")

	flag.StringVar(&ret.Log, "logfile", "gamma.log",
		"Log messages will be saved here.")
	
    flag.StringVar(&ret.CachePath, "cache", DefaultCachePath,
		"Path to cached files.")

	flag.BoolVar(&ret.Skip, "skip_optional", false,
		"If set the proccessing will skip optional steps.")
	flag.BoolVar(&ret.Show, "show_steps", false,
		"If set, prints the processing steps.")
    
    err = flag.Parse(args)
    
    if err != nil {
        return
    }
    
	return ret, nil
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

func (self *Process) Parse() (istart int, istop int, err error) {
	handle := Handler("Process.Parse")
    
	if self.Show {
		listSteps()
		os.Exit(0)
	}

    istep, istart, istop := stepIndex(self.Step), stepIndex(self.Start),
		stepIndex(self.Stop)

	if istep == -1 {
		if istart == -1 {
			listSteps()
            err = handle(nil,
                "Starting step '%s' is not in list of available steps!",
                self.Start)
            return
		}

		if istop == -1 {
			listSteps()
            err = handle(nil,
                "Stopping step '%s' is not in list of available steps!",
                self.Stop)
			return
		}
	} else {
		istart = istep
		istop = istep + 1
	}

	path := self.Conf
	data, err := ReadFile(path)

	if err != nil {
		err = handle(err, "Failed to read file:  '%s'!", path)
        return
	}

	if err = json.Unmarshal(data, &self.config); err != nil {
		err = handle(err, "Failed to parse json data: %s'!", data)
        return
	}

	return istart, istop, nil
}

func (self *Process) RunSteps(start, stop int) error {
	handle := Handler("Process.RunSteps")

	for ii := start; ii < stop; ii++ {
		name := stepList[ii]
		step, _ := steps[name]

		delim(fmt.Sprintf("START: %s", name), "*")

		if err := step(&self.config); err != nil {
			return handle(err, "Error while running step: '%s'",
				name)
		}

		delim(fmt.Sprintf("END: %s", name), "*")
	}
	return nil
}


func InitParse(args []string) (ret string, err error) {
    flag := fl.NewFlagSet("init", fl.ContinueOnError)
    
	flag.StringVar(&ret, "config", "gamma.json",
		"Processing configuration file")
    
    err = flag.Parse(args)
    
    if err != nil {
        return
    }
    
	return ret, nil
}


func NewLister(args []string) (ret Lister, err error) {
    handle := Handler("NewLister")
    flag := fl.NewFlagSet("init", fl.ContinueOnError)
    
	flag.StringVar(&ret.conf, "config", "gamma.json",
		"Processing configuration file")
    flag.StringVar(&ret.Mode, "mode", "", "Which type of files to list.")
    
    err = flag.Parse(args)
    
    if err != nil {
        return
    }
    
    path := ret.conf
    
    err = LoadJson(path, &ret.config)
    
    
    if err != nil {
        err = handle(err, "Failed to parse json file: '%s'!", path)
        return
    }
    
    path = ret.General.Metafile
    
    err = LoadJson(path, &ret.Meta)
    
    if err != nil {
        err = handle(err, "Failed to parse json file: '%s'!",
            path)
        return
    }
    
	return ret, nil
}


func (self *Lister) Quicklook() error {
    handle := Handler("Lister.Quicklook")
    
    cache := fp.Join(self.General.CachePath, "sentinel1")
    
    info := &ExtractOpt{root:cache, pol:self.General.Pol}
    
    zips := self.Meta.Zipfiles
    
    for _, zip := range zips {
        s1, err := NewS1Zip(zip, cache)
        
        if err != nil {
            return handle(err, 
                "Failed to parse Sentinel-1 information from zipfile '%s'!",
                s1.Path)
        }
        
        image, err := s1.Quicklook(info)
        
        if err != nil {
            return handle(err, "Failed to retreive quicklook file in zip '%s'!",
                s1.Path)
        }
        
        fmt.Println(image)
    } 
    
    return nil
}