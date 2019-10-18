package gamma

import (
    "log"
    "encoding/json"
    "fmt"
    "os"
    fl "flag"
    fp "path/filepath"
    str "strings"
)

type (
    Process struct {
        Conf        string `name:"conf" default:"gamma.conf"`
        Step        string `name:"step" default:""`
        Start       string `name:"start" default:""`
        Stop        string `name:"stop" default:""` 
        Log         string `name:"log" default:""`
        CachePath   string `name:"cache" default:""`
        Skip        bool   `name:"skip" default:""`
        Show        bool   `name:"show" default:""`
        Config
    }

    Batcher struct {
        Conf     string `name:"conf" default:"gamma.conf"`
        Mode     string `name:"" default:""`
        Infile   string `name:"file" default:""`
        OutDir   string `name:"out" default:"."`
        Filetype string `name:"ftype" default:""`
        Mli      string `name:"mli" default:""`
        Config
        ifgPlotOpt
    }

    Displayer struct {
        Dat  string `pos:"1"`
        Par  string `name:"par" default:""`
        Mode string `pos:"0"`
        Sec  string `name:"sec" default:""`
        rasArgs
    }
    
    Coder struct {
        GeoMeta
        CodeOpt
        infile   string `pos:"0"`
        outfile  string `name:"out" default:""`
        metafile string `name:"meta" default:"geocode.json"`
        mode     string `name:"mode" default:""`
    }
)

var (
    BatchModes = []string{"quicklook", "mli / MLI", "ras"}
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

    if err = json.Unmarshal(data, &proc.Config); err != nil {
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

        if err := step(&proc.Config); err != nil {
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

func addRasArgsFlags(args *rasArgs, flag *fl.FlagSet) {
    flag.IntVar(&args.Rng, "rng", 0, "Range samples of datafile.")
    flag.IntVar(&args.Azi, "Azi", 0, "Azimuth lines of datafile.")
    
    flag.BoolVar(&args.Flip, "flip", false,
        "Should the output image be flipped.")
    
    flag.StringVar(&args.Cmd, "cmd", "", "Plot command type to be used.")
    
    flag.IntVar(&args.Start, "start", 0, "Starting lines.")
    flag.IntVar(&args.Nlines, "nline", 0, "Number of lines to plot.")
    
    flag.Float64Var(&args.Scale, "scale", 1.0, "Display scale factor.")
    flag.Float64Var(&args.Exp, "exp", 0.35, "Display exponent.")
    
    flag.IntVar(&args.avgFact, "avg", 1000, "Averaging factor of pixels.")
    flag.IntVar(&args.headerSize, "header", 0, "Header size?.")
}


func NewBatcher(args []string) (ret Batcher, err error) {
    flag := fl.NewFlagSet("list", fl.ContinueOnError)
    
    if len(args) == 0 {
        err = Handle(nil, "at least one argument is required")
        return
    }
    
    ret.Mode = args[0]

    flag.StringVar(&ret.Conf, "config", "gamma.json",
        "Processing configuration file")
    flag.StringVar(&ret.Infile, "file", "", "Inputfile.")
    flag.StringVar(&ret.OutDir, "out", ".", "Output directory.")
    flag.StringVar(&ret.Filetype, "ftype", "", "Type of files located in infile.")
    
    addRasArgsFlags(&ret.ifgPlotOpt.rasArgs, flag)
    
    flag.StringVar(&ret.Mli, "mli", "", "MLI datafile used for background.")
    flag.IntVar(&ret.startCC, "startCC", 1, "Start coherence lines.")
    flag.IntVar(&ret.startPwr, "startPwr", 1, "Start power lines.")
    flag.IntVar(&ret.startCpx, "startCpx", 1, "Start complex lines.")
    flag.Float64Var(&ret.Range.Min, "min", 0.0, "Minimum value.")
    flag.Float64Var(&ret.Range.Max, "max", 0.0, "Maximum value.")
    
    err = flag.Parse(args[1:])

    if err != nil {
        return
    }

    if len(ret.infile) == 0 {
        err = Handle(nil, "inputfile not specified")
        return
    }

    path := ret.Conf
    err = LoadJson(path, &ret.Config)

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
    path := b.infile
    file, err := NewReader(path)

    if err != nil {
        return Handle(err, "failed to create FileReader '%s'!", path)
    }

    defer file.Close()
    
    opt := &MLIOpt {
        Looks: b.General.Looks,
    }
    
    mliDir := b.OutDir
    
    err = os.MkdirAll(mliDir, os.ModePerm)
    
    if err != nil {
        return Handle(err, "failed to create directory '%s'", mliDir)
    }
    
    for file.Scan() {
        line := file.Text()

        s1, err := FromTabfile(line)

        if err != nil {
            return Handle(err, "failed to parse S1SLC tabfile '%s'", line)
        }
        
        dat := fp.Join(mliDir, s1.Format(DateShort) + ".mli")
        mli, err := NewMLI(dat, "")
        
        if err != nil {
            return Handle(err, "could not create MLI struct")
        }
        
        exist, err := mli.Exist()
        
        if err != nil {
            return Handle(err, "failed to check whether MLI file exists")
        }
        
        if exist {
            fmt.Printf("%s %s\n", mli.Dat, mli.Par)
            continue
        }
        
        err = s1.MLI(&mli, opt)
        
        if err != nil {
            return Handle(err, "failed to retreive MLI file for '%s'",
                line)
        }

        fmt.Printf("%s %s\n", mli.Dat, mli.Par)
    }

    return nil
}


func (b *Batcher) Raster() error {
    path := b.infile
    file, err := NewReader(path)

    if err != nil {
        return Handle(err, "failed to create FileReader '%s'!", path)
    }

    defer file.Close()
    
    switch b.Filetype {
    case "mli", "MLI":
        err := b.PlotMLIs(&file)
        
        if err != nil {
            return Handle(err, "plotting of MLI datafiles failed")
        }
    case "slc", "SLC":
        err := b.PlotSLCs(&file)
        
        if err != nil {
            return Handle(err, "plotting of SLC datafiles failed")
        }
    case "ifg", "IFG":
        err := b.PlotIFGs(&file)
        
        if err != nil {
            return Handle(err, "plotting of IFG datafiles failed")
        }
    default:
        return fmt.Errorf("unrecognized filetype '%s'", b.Filetype) 
    }
    
    return nil
}


func NewDisplayer(args []string) (ret Displayer, err error) {
    flag := fl.NewFlagSet("display", fl.ContinueOnError)

    ret.Mode = args[0]
    
    flag.StringVar(&ret.Dat, "dat", "",
        "Datafile containing data to plot.")
    flag.StringVar(&ret.Par, "par", "", "Parfile describing datafile.")
    
    flag.StringVar(&ret.Sec, "sec", "", "Secondary input datafile.")
    
    addRasArgsFlags(&ret.rasArgs, flag)    
    
    err = flag.Parse(args[1:])
    
    if err != nil {
        err = Handle(err, "failed to parse command line options")
        return
    }
    
    if len(ret.Dat) == 0 {
        err = Handle(nil, "dat should be valied path not empty string")
        return
    }
    
    split := str.Split(ret.Dat, ".")
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
    dat, err := NewDataFile(dis.Dat, dis.Par, "par")
    
    if err != nil {
        return Handle(err, "failed to parse datafile '%s'", dis.Dat)
    }
    
    err = dis.disArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse plotting options")
    }
    
    switch dis.Mode {
    case "dis":
        err := Display(dat,  dis.rasArgs.disArgs)
        
        if err != nil {
            return Handle(err, "failed to execute display")
        }
    
    case "ras":
        err := Raster(dat,  dis.rasArgs, dis.Sec)
        
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
    return g.Dem.geo2radar(g.infile, g.outfile, g.CodeOpt)
}

// geocode_back = radar2geo

func (g *Coder) Radar2Geo() error {
    return g.Dem.radar2geo(g.infile, g.outfile, g.CodeOpt)
}

var PlotCmdFiles = map[string]Slice{
    "pwr": Slice{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": Slice{"slc", "rslc"},
    "mph": Slice{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": Slice{"hgt", "rdc"},
}

func (b *Batcher) PlotMLIs(file *FileReader) error {
    opt := b.ifgPlotOpt.rasArgs
    
    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := str.Fields(line)
        
        log.Printf("Processing: %s\n", line)
        
        mli, err := NewMLI(split[0], split[1])
        
        if err != nil {
            return Handle(err, "failed to parse MLI file from line '%s'", line)
        }
            
        err = mli.Raster(opt)
        
        if err != nil {
            return Handle(err, "failed to create raster for '%s'", line)
        }
    }
    return nil
}

func (b *Batcher) PlotSLCs(file *FileReader) error {
    opt := b.ifgPlotOpt.rasArgs
    
    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := str.Fields(line)
        
        log.Printf("Processing: %s\n", line)
        
        slc, err := NewSLC(split[0], split[1])
        
        if err != nil {
            return Handle(err, "failed to parse SLC file from line '%s'", line)
        }
            
        err = slc.Raster(opt)
        
        if err != nil {
            return Handle(err, "failed to create raster for '%s'", line)
        }
    }
    return nil
}

func (b *Batcher) PlotIFGs(file *FileReader) error {
    opt := b.ifgPlotOpt
    mli := b.Mli
    
    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := str.Fields(line)
        
        log.Printf("Processing: %s\n", line)
        
        dat := split[0]
        par, diffPar, simUnwrap, quality := "", "", "", ""
        
        if len(split) > 2 {
            par = split[1]
        }
        
        if len(split) > 3 {
            simUnwrap = split[2]
        }
        
        if len(split) > 4 {
            diffPar = split[3]
        }
        
        if len(split) > 5 {
            quality = split[4]
        }
        
        ifg, err := NewIFG(dat, par, simUnwrap, diffPar, quality)
        
        if err != nil {
            return Handle(err, "failed to parse IFG file from line '%s'", line)
        }
            
        err = ifg.Raster(mli, opt)
        
        if err != nil {
            return Handle(err, "failed to create raster for '%s'", line)
        }
    }
    return nil
}
