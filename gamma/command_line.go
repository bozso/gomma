package gamma

import (
	//"log"
	"encoding/json"
	fl "flag"
	"fmt"
	"os"
	fp "path/filepath"
	//str "strings"
)

type (
	Process struct {
		Conf, Step, Start, Stop, Log, CachePath string
		Skip, Show                              bool
		config
	}

	Lister struct {
		conf, Mode, infile string
		config
	}

	Meta struct {
		MasterIdx  int
		MasterDate string
	}
)

var (
	ListModes = []string{"quicklook"}
)

func NewProcess(args []string) (ret Process, err error) {
	flag := fl.NewFlagSet("proc", fl.ContinueOnError)

	flag.StringVar(&ret.Conf, "config", "gamma.json",
		"Processing configuration file")

	flag.StringVar(&ret.infile, "file", "",
		"Infile. List of files to process.")

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

func (proc *Process) Parse() (istart int, istop int, err error) {
	if proc.Show {
		listSteps()
		os.Exit(0)
	}

	istep, istart, istop := stepIndex(proc.Step), stepIndex(proc.Start),
		stepIndex(proc.Stop)

	if istep == -1 {
		if istart == -1 {
			listSteps()
			err = Handle(nil,
				"Starting step '%s' is not in list of available steps!",
				proc.Start)
			return
		}

		if istop == -1 {
			listSteps()
			err = Handle(nil,
				"Stopping step '%s' is not in list of available steps!",
				proc.Stop)
			return
		}
	} else {
		istart = istep
		istop = istep + 1
	}

	path := proc.Conf
	data, err := ReadFile(path)

	if err != nil {
		err = Handle(err, "Failed to read file:  '%s'!", path)
		return
	}

	if err = json.Unmarshal(data, &proc.config); err != nil {
		err = Handle(err, "Failed to parse json data: %s'!", data)
		return
	}

	return istart, istop, nil
}

func (proc *Process) RunSteps(start, stop int) error {
	for ii := start; ii < stop; ii++ {
		name := stepList[ii]
		step, _ := steps[name]

		delim(fmt.Sprintf("START: %s", name), "*")

		if err := step(&proc.config); err != nil {
			return Handle(err, "Error while running step: '%s'",
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
	flag := fl.NewFlagSet("init", fl.ContinueOnError)

	mode := args[0]

	if mode != "quicklook" {
		err = Handle(nil, "Unrecognized lister mode '%s'", mode)
		return
	}

	ret.Mode = mode

	flag.StringVar(&ret.conf, "config", "gamma.json",
		"Processing configuration file")
	flag.StringVar(&ret.infile, "file", "", "Inputfile.")

	err = flag.Parse(args[1:])

	if err != nil {
		return
	}

	if len(ret.infile) == 0 {
		err = Handle(nil, "Inputfile must by specified!")
		return
	}

	path := ret.conf
	err = LoadJson(path, &ret.config)

	if err != nil {
		err = Handle(err, "Failed to parse json file: '%s'!", path)
		return
	}

	return ret, nil
}

func (self *Lister) Quicklook() error {
	cache := fp.Join(self.General.CachePath, "sentinel1")

	info := &ExtractOpt{root: cache, pol: self.General.Pol}

	path := self.infile
	file, err := NewReader(path)

	if err != nil {
		return Handle(err, "Could not create FileReader for file '%s'!", path)
	}

	defer file.Close()

	for file.Scan() {
		line := file.Text()

		s1, err := NewS1Zip(line, cache)

		if err != nil {
			return Handle(err,
				"Failed to parse Sentinel-1 information from zipfile '%s'!",
				s1.Path)
		}

		image, err := s1.Quicklook(info)

		if err != nil {
			return Handle(err, "Failed to retreive quicklook file in zip '%s'!",
				s1.Path)
		}

		fmt.Println(image)
	}

	return nil
}
