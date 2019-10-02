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

    Batcher struct {
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
    
    Coder struct {
        GeoMeta
        CodeOpt
        infile, outfile, metafile, mode string
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

func NewBatcher(args []string) (ret Batcher, err error) {
    flag := fl.NewFlagSet("list", fl.ContinueOnError)
    
    if len(args) == 0 {
        err = Handle(nil, "at least one argument is required")
        return
    }
    
    mode := args[0]

    if mode != "quicklook" && mode != "mli" {
        err = Handle(nil, "unrecognized Batcher mode '%s'", mode)
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

func (b *Batcher) Quicklook() error {
    cache := fp.Join(b.General.CachePath, "sentinel1")

    info := &ExtractOpt{root: cache, pol: b.General.Pol}

    path := b.infile
    file, err := NewReader(path)

    if err != nil {
        return Handle(err, "failed to create FileReader '%s'!", path)
    }

    defer file.Close()

    for file.Scan() {
        line := file.Text()

        s1, err := NewS1Zip(line, cache)

        if err != nil {
            return Handle(err, "failed to parse Sentinel-1 '%s'", s1.Path)
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

func (b *Batcher) MLI() error {
    root := fp.Join(b.General.CachePath, "sentinel1")

    path, pol := b.infile, b.General.Pol
    file, err := NewReader(path)

    if err != nil {
        return Handle(err, "failed to create FileReader '%s'!", path)
    }

    defer file.Close()
    
    opt := &MLIOpt {
        Looks: b.General.Looks,
    }

    for file.Scan() {
        line := file.Text()

        s1, err := NewS1Zip(line, root)

        if err != nil {
            return Handle(err, "failed to parse Sentinel-1 '%s'", s1.Path)
        }
        
        dat, par := s1.Names("mli", pol)
        mli, err := NewMLI(dat, par)
        
        if err != nil {
            return Handle(err, "could not create MLI struct")
        }
        
        exist, err := mli.Exist()
        
        if err != nil {
            return Handle(err, "failed to check whether MLI file exists")
        }
        
        if exist {
            fmt.Printf("%s %s\n", mli.dat, mli.par)
            continue
        }
        
        slc, err := s1.SLC(pol)
        
        if err != nil {
            return Handle(err, "failed to create S1SLC struct")
        }
        
        err = slc.MLI(&mli, opt)
        
        if err != nil {
            return Handle(err, "failed to retreive MLI file for '%s'",
                s1.Path)
        }

        fmt.Printf("%s %s\n", mli.dat, mli.par)
    }

    return nil
}


func NewDisplayer(args []string) (ret Displayer, err error) {
    flag := fl.NewFlagSet("display", fl.ContinueOnError)

    ret.mode = args[0]
    
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
    dat, err := NewDataFile(dis.dat, dis.par, "par")
    
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


func NewCoder(args []string) (ret Coder, err error) {
    flag := fl.NewFlagSet("coding", fl.ContinueOnError)
    
    flag.StringVar(&ret.infile, "in", "", "Datafile containing to transform.")
    flag.StringVar(&ret.outfile, "out", "", "Output datafile.")
    flag.StringVar(&ret.metafile, "meta", "geocode.json",
        "Metadata of geocoding.")
    
    flag.IntVar(&ret.inWidth, "inwidth", 0,
        "Range samples or Width of infile.")
    flag.IntVar(&ret.outWidth, "outwidth", 0,
        "Range samples or Width of outfile.")
    
    flag.IntVar(&ret.nlines, "nlines", 0, "Number of lines to code.")
    flag.StringVar(&ret.mode, "intpol", "nearest", "Interpolation mode.")
    flag.IntVar(&ret.order, "order", 0,
        "Lanczos function order or B-spline degree.")
    flag.BoolVar(&ret.flipInput, "flipIn", false, "Flip input.")
    flag.BoolVar(&ret.flipOutput, "flipOut", false, "Flip output.")
    
    err = flag.Parse(args[1:])
    
    if err != nil {
        err = Handle(err, "failed to parse command line options")
        return
    }
    
    switch ret.mode {
    case "nearest":
        ret.interpolMode = NearestNeighbour
    case "bic":
        ret.interpolMode = BicubicSpline
    case "bic_log":
        ret.interpolMode = BicubicSplineLog
    case "bic_sqrt":
        ret.interpolMode = BicubicSplineSqrt
    case "bsp":
        ret.interpolMode = BSpline
    case "bsp_sqrt":
        ret.interpolMode = BSplineSqrt
    case "lanc":
        ret.interpolMode = Lanczos
    case "lanc_sqrt":
        ret.interpolMode = LanczosSqrt
    case "inv_dist":
        ret.interpolMode = InvDist
    case "inv_sqrd_dist":
        ret.interpolMode = InvSquaredDist
    case "const":
        ret.interpolMode = Constant
    case "gauss":
        ret.interpolMode = Gauss
    default:
        err = Handle(nil, "unrecognized interpolation mode '%s'", ret.mode)
        return
    }
    
    return ret, nil
}

// geocode = geo2radar

func (g *Coder) Geo2Radar() error {
    return g.dem.geo2radar(g.infile, g.outfile, g.CodeOpt)
}

// geocode_back = radar2geo

func (g *Coder) Radar2Geo() error {
    return g.dem.radar2geo(g.infile, g.outfile, g.CodeOpt)
}

var PlotCmdFiles = map[string]Slice{
    "pwr": Slice{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": Slice{"slc", "rslc"},
    "mph": Slice{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": Slice{"hgt", "rdc"},
}
