package gamma

import (
    //"log"
    "encoding/json"
    "fmt"
    "os"
    fl "flag"
    fp "path/filepath"
    str "strings"
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
    
    Displayer struct {
        dat, par, mode, sec string
        rasArgs
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
        err = Handle(err, "NewProcess failed")
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
                "start step '%s' not in list of available steps!",
                proc.Start)
            return
        }

        if istop == -1 {
            listSteps()
            err = Handle(nil,
                "stop step '%s' not in list of available steps!",
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
        err = Handle(err, "failed to read file '%s'", path)
        return
    }

    if err = json.Unmarshal(data, &proc.config); err != nil {
        err = Handle(err, "failed to parse json data '%s'", data)
        return
    }

    return istart, istop, nil
}

func (proc *Process) RunSteps(start, stop int) error {
    for ii := start; ii < stop; ii++ {
        name := stepList[ii]
        step := steps[name]

        delim(fmt.Sprintf("START: %s", name), "*")

        if err := step(&proc.config); err != nil {
            return Handle(err, "error while running step '%s'",
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
    flag := fl.NewFlagSet("list", fl.ContinueOnError)

    mode := args[0]

    if mode != "quicklook" {
        err = Handle(nil, "unrecognized lister mode '%s'", mode)
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
        err = Handle(nil, "inputfile not specified")
        return
    }

    path := ret.conf
    err = LoadJson(path, &ret.config)

    if err != nil {
        err = Handle(err, "failed to parse json file '%s'", path)
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
        return Handle(err, "failed to create FileReader '%s'!", path)
    }

    defer file.Close()

    for file.Scan() {
        line := file.Text()

        s1, err := NewS1Zip(line, cache)

        if err != nil {
            return Handle(err,
                "failed to parse Sentinel-1 '%s'",
                s1.Path)
        }

        image, err := s1.Quicklook(info)

        if err != nil {
            return Handle(err, "failed to retreive quicklook file '%s'",
                s1.Path)
        }

        fmt.Println(image)
    }

    return nil
}

func NewDisplayer(args []string) (ret Displayer, err error) {
    flag := fl.NewFlagSet("display", fl.ContinueOnError)

    mode := args[0]

    if mode != "ras" && mode != "dis" {
        err = Handle(nil, "unrecognized display mode '%s'", mode)
        return
    }

    ret.mode = mode
    
    flag.StringVar(&ret.dat, "dat", "",
        "Datafile containing data to plot.")
    flag.StringVar(&ret.par, "par", "", "Parfile describing datafile.")
    
    flag.StringVar(&ret.sec, "sec", "", "Secondary input datafile.")
    
    flag.IntVar(&ret.Rng, "rng", 0, "Range samples of datafile.")
    flag.IntVar(&ret.Azi, "Azi", 0, "Azimuth lines of datafile.")
    
    flag.BoolVar(&ret.Flip, "flip", false,
        "Should the output image be flipped.")
    
    flag.StringVar(&ret.Cmd, "cmd", "", "Plot command type to be used.")
    
    flag.IntVar(&ret.Start, "start", 0, "Starting lines.")
    flag.IntVar(&ret.Nlines, "nline", 0, "Number of lines to plot.")
    
    flag.Float64Var(&ret.Scale, "scale", 1.0, "Display scale factor.")
    flag.Float64Var(&ret.Exp, "exp", 0.35, "Display exponent.")
    
    flag.IntVar(&ret.rasArgs.avgFact, "avg", 1000, "Averaging factor of pixels.")
    flag.IntVar(&ret.rasArgs.headerSize, "header", 0, "Header size?.")
    
    err = flag.Parse(args[1:])
    
    if err != nil {
        err = Handle(err, "failed to parse command line options")
        return
    }
    
    if len(ret.dat) == 0 {
        err = Handle(nil, "dat should be valied path not empty string")
        return
    }
    
    split := str.Split(ret.dat, ".")
    ext := split[len(split)-1]
    
    if len(ret.Cmd) == 0 {
        for key, val := range PlotCmdFiles {
            if val.Contains(ext) {
                ret.Cmd = key
            }
        }
        
        if len(ret.Cmd) == 0 {
            err = Handle(nil,
                "could not determine plot command from extension '%s'",
                ext)
            return
        }
    }
    
    return ret, nil
}

func (dis *Displayer) Plot() error {
    dat, err := NewDataFile(dis.dat, dis.par)
    
    if err != nil {
        return Handle(err, "failed to parse datafile '%s'", dis.dat)
    }
    
    err = dis.disArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse plotting options")
    }
    
    switch dis.mode {
    case "dis":
        err := Display(dat,  dis.rasArgs.disArgs)
        
        if err != nil {
            return Handle(err, "failed to execute display")
        }
    
    case "ras":
        err := Raster(dat,  dis.rasArgs, dis.sec)
        
        if err != nil {
            return Handle(err, "failed to execute raster")
        }
    }
    return nil
}

var PlotCmdFiles = map[string]Slice{
    "pwr": Slice{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": Slice{"slc", "rslc"},
    "mph": Slice{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": Slice{"hgt", "rdc"},
}
