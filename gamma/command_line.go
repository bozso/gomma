package gamma

import (
    "log"
    "encoding/json"
    "fmt"
    "os"
    fp "path/filepath"
    //ref "reflect"
    str "strings"
)

type (
    Cmd func(Args) error
    JSONMap map[string]interface{}
    
    Process struct {
        Conf        string `name:"conf" default:"gamma.conf"`
        Step        string `name:"step" default:""`
        Start       string `name:"start" default:""`
        Stop        string `name:"stop" default:""` 
        Log         string `name:"log" default:"gamma.log"`
        CachePath   string `name:"cache" default:""`
        Skip        bool   `name:"skip"`
        Show        bool   `name:"show"`
        Config
    }
    
    MetaFile struct {
        Meta string `pos:"0"`
    }
)

const (
    ParseErr CWerror = "failed to parse command line arguments"
)

var (
    BatchModes = []string{"quicklook", "mli / MLI", "ras"}
    
    Commands = map[string]Cmd{
        "proc": process,
        "init": initGamma,
        "batch": batch,
        "move": move,
        "make": create,
        "stat": stat,
        //"splitIfg": splitIfg,
    }
    
    CommandsAvailable = MapKeys(Commands)
)

func process(args Args) (err error) {
    proc := Process{}
    
    if err = args.ParseStruct(&proc); err != nil {
        err = ParseErr.Wrap(err)
        //err = Handle(err, parseErr)
        return
    }
    
    start, stop, err := proc.Parse()
    if err != nil {
        err = Handle(err, "failed to  parse processing steps")
        return
    }
    
    if err = proc.RunSteps(start, stop); err != nil {
        err = Handle(err, "error occurred while running processing steps")
        return
    }
    return nil
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


const (
    StepErr Werror = "start step '%s' not in list of available steps"
)


func (proc Process) Parse() (istart int, istop int, err error) {
    if proc.Show {
        listSteps()
        os.Exit(0)
    }

    istep, istart, istop := stepIndex(proc.Step), stepIndex(proc.Start),
        stepIndex(proc.Stop)

    if istep == -1 {
        if istart == -1 {
            listSteps()
            err = StepErr.Make(proc.Start)
            //err = Handle(nil,
                //"start step '%s' not in list of available steps!",
                //proc.Start)
            return
        }

        if istop == -1 {
            listSteps()
            err = StepErr.Make(proc.Stop)
            //err = Handle(nil,
                //"stop step '%s' not in list of available steps!",
                //proc.Stop)
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

func (proc Process) RunSteps(start, stop int) error {
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


func initGamma(args Args) (err error) {
    if len(args.pos) < 1 {
        err = Handle(nil,
            "at least one positional argument required (path of configuration file")
        return
    }
    
    conf := args.pos[0]
    
    if err = MakeDefaultConfig(conf); err != nil {
        err = Handle(err, "failed to create config file")
        return
    }
    return
}

type(
    Batcher struct {
        Mode     string `pos:"0"`
        Infile   string `pos:"1"`
        Conf     string `name:"conf" default:"gamma.conf"`
        OutDir   string `name:"out" default:"."`
        Filetype string `name:"ftype" default:""`
        Mli      string `name:"mli" default:""`
        RasArgs
        Config
    }
)

func batch(args Args) (err error) {
    batch := Batcher{}
    
    if err = args.ParseStruct(&batch); err != nil {
        err = ParseErr.Wrap(err)
        //err = Handle(err, parseErr)
        return
    }

    //fmt.Printf("%#v\n", batch)
    
    switch batch.Mode {
    case "quicklook":
        if err = batch.Quicklook(); err != nil {
            err = Handle(err, "quicklook failed")
            return
        }
    case "mli", "MLI":
        if err = batch.MLI(); err != nil {
            err = Handle(err, "MLI failed")
            return
        }
    case "ras", "raster", "plot":
        if err = batch.Raster(); err != nil {
            err = Handle(err, "raster failed")
            return
        }
    default:
        err = Handle(nil, "unrecognized mode: '%s'! Choose from: %v",
            batch.Mode, BatchModes)
        return
    }
    return nil
}

func (b Batcher) Quicklook() error {
    cache := fp.Join(b.General.CachePath, "sentinel1")

    info := &ExtractOpt{root: cache, pol: b.General.Pol}

    path := b.infile
    file, err := NewReader(path)

    if err != nil {
        return err
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

func (b Batcher) MLI() error {
    path := b.infile
    file, err := NewReader(path)

    if err != nil {
        return err
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
            return err
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

func (b Batcher) Raster() error {
    path := b.Infile
    opt := b.RasArgs
    
    file, err := NewReader(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        var data DatFile
        if err = Load(line, &data); err != nil {
            return Handle(err, "failed to load datafile from '%s'", line)
        }
        
        if err = data.Raster(opt); err != nil {
            return Handle(err, "failed to create raster file for '%s'", line)
        }
    }
    
    return nil
}

type(
    Mover struct {
        OutDir   string `name:"out" default:"."`
        MetaFile
    }
)

func move(args Args) (err error) {
    m := Mover{}
    
    if err = args.ParseStruct(&m); err != nil {
        err = ParseErr.Wrap(err)
        return
    }
    
    path := m.Meta
    
    dat, err := LoadDataFile(path)
    if err != nil {
        err = Handle(err, "failed to parse json metadatafile '%s'", path)
        return
    }
    
    log.Fatalf("%s\n", dat.TypeStr())
    
    out := m.OutDir
    
    if dat, err = dat.Move(out); err != nil {
        //err = Handle(err, "failed to move datafile to '%s", out)
        return err
    }
    
    if path, err = Move(path, out); err != nil {
        //err = Handle(err, "failed to move json metafile to '%s'", out)
        return err
    }
    
    if err = SaveJson(path, dat); err != nil {
        err = Handle(err, "failed to refresh json metafile")
        return err
    }
    
    return nil
}

type (
    Coregister struct {
        CoregOpt
        Master  string `name:"master" default:""`
        Slave   string `name:"slave" default:""`
        Ref     string `name:"ref" default:""`
        Outfile string `name:"out" default:""`
    }

)

/*
func coreg(args Args) error {
    c := Coregister{}
    
    args.ParseStruct(&c)
    
    sm, ss, sr := c.Master, c.Slave, c.Ref
    
    if len(sm) == 0 {
        return EmptyStringErr.Make("master", sm)
    }
    
    if len(ss) == 0 {
        return EmptyStringErr.Make("slave", ss)
    }
    
    ref *S1SLC
    
    if len(r) == 0 {
        ref = nil
    } else {
        r, err := FromTabfile(sr)
        if err != nil {
            return ParseTabErr.Wrap(sr)
        }
        ref = &r
    }
    
    s, err := FromTabfile(ss)
    if err != nil {
        return ParseTabErr.Wrap(ss)
    }
    
    m, err := FromTabfile(sm)
    if err != nil {
        return ParseTabErr.Wrap(sm)
    }
    
    
    c.CoregOpt
}
*/

type Creater struct {
    Dtype      string `name:"dtype"`
    Ftype      string `name:"ftype"`
    ParfileExt string `name:"parExt"`
    DatParFile
    MetaFile
}

func create(args Args) (err error) {
    c := Creater{}
    
    if err = args.ParseStruct(&c); err != nil {
        err = ParseErr.Wrap(err)
        return
    }
    
    dat := c.Dat
    
    if len(dat) == 0 {
        if err = c.DatParFile.Tmp(); err != nil {
            return
        }
    }
    
    ext := c.ParfileExt
    if len(ext) == 0 {
        ext = "par"
    }
    
    par := c.Par
    if len(par) == 0 {
        par = fmt.Sprintf("%s.%s", dat, ext)
    }
    
    
    dt_, dt := c.Dtype, Unknown
    
    if len(dt_) > 0 {
        dt, err = str2dtype(dt_)
        if err != nil {
            return
        }
    }
    
    dat, err = fp.Abs(dat)
    if err != nil { return }
    
    par, err = fp.Abs(par)
    if err != nil { return }
        
    datf, err := NewDataFile(dat, par, dt) 
    if err != nil {
        err = DataCreateErr.Wrap(err, "DataFile")
        return
    }
    
    
    var Dat DataFile
    ftype := str.ToUpper(c.Ftype)
    switch ftype {
    case "MLI":
        Dat = MLI{dataFile: datf}
    default:
        err = fmt.Errorf("unrecognized filetype '%s'", ftype)
        return
    }
    
    if err = Dat.Save(c.Meta); err != nil {
        err = Handle(err, "failed to save datafile to metafile '%s'",
            c.MetaFile)
        return
    }
    
    return nil
}

/*
type (
    SplitIfg struct {
        SBIOpt
        SSIOpt
        SpectrumMode string `name:"mode"`
        Master       string `name:"master"`
        Slave        string `name:"slave"`
        Mli          string `name:"mli"`
    }
)

func splitIfg(args Args) error {
    si := SplitIfg{}
    si.Mode = Ifg
    
    if err := args.ParseStruct(&si); err != nil {
        return ParseErr.Wrap(err)
    }
    
    ms, ss := si.Master, si.Slave
    
    if len(ms) == 0 || len(ss) == 0 {
        return fmt.Errorf("both master and slave SLC files should be specified")
    }
    
    var (
        m_, s_ DataFile
        m, s SLC
        err error
    )
    
    if m_, err := LoadDataFile(ms); err != nil {
        return err
    }
    
    if m, ok := m_.(SLC); !ok {
        return TypeErr.Make(m, "master", "SLC")
    }
    
    if s_, err = LoadDataFile(ms); err != nil {
        return err
    }
    
    if s, ok := s_.(SLC); !ok {
        return TypeErr.Make(m, "slave", "SLC")
    }
    
    id := ID(m, s, DShort)
    
    mode := str.ToUpper(si.SpectrumMode)
    
    switch mode {
    case "BEAM", "B":
        opt := si.SBIOpt
        
        out, err := m.SplitBeamIfg(s, opt)
        
        if err != nil {
            return err
        }
        
        if err = out.Mli.Save(id + "_sbi_mli.json"); err != nil {
            return err
        }
        
        if err = out.Ifg.Save(id + "_sbi_ifg.json"); err != nil {
            return err
        }
    case "SPECTRUM", "S":
        opt := si.SSIOpt
        
        Mli, err := LoadDataFile(si.Mli)
        if err != nil {
            return err
        }
        
        mli, ok := Mli.(MLI)
        
        if !ok {
            return TypeErr.Make(Mli, "mli", "MLI")
        }
        
        out, err := m.SplitSpectrumIfg(s, mli, opt)
        
        if err != nil {
            return err
        }
        
        // still need to figure out the returned files
        return nil
    default:
        return fmt.Errorf("unrecognized Split Interferogram mode: '%s'", mode)
    }
    return nil
}
*/

type (
    Stat struct {
        Out string `pos:"1"`
        MetaFile
        Subset
    }
)

var imgStat = Gamma.Must("image_stat")

func stat(args Args) (err error) {
    s := Stat{}
    
    if err := args.ParseStruct(&s); err != nil {
        return ParseErr.Wrap(err)
    }
    
    var dat DatFile
    
    if err = Load(s.Meta, &dat); err != nil {
        return
    }
    
    //s.Subset.Parse(dat)
    
    _, err = imgStat(dat.Datfile(), dat.Rng(), s.RngOffset, s.AziOffset,
                     s.RngWidth, s.AziLines, s.Out)
    
    return
}

/*

type(
    Displayer struct {
        Mode string `pos:"0"`
        Dat  string `pos:"1"`
        Par  string `name:"par" default:""`
        Sec  string `name:"sec" default:""`
        RasArgs
    }
)

func NewDisplayer(args []string) (ret Displayer, err error) {
    flag := fl.NewFlagSet("display", fl.ContinueOnError)

    ret.Mode = args[0]
    
    flag.StringVar(&ret.Dat, "dat", "",
        "Datafile containing data to plot.")
    flag.StringVar(&ret.Par, "par", "", "Parfile describing datafile.")
    
    flag.StringVar(&ret.Sec, "sec", "", "Secondary input datafile.")
    
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
    
    err = dis.DisArgs.Parse(dat)
    
    if err != nil {
        return Handle(err, "failed to parse plotting options")
    }
    
    switch dis.Mode {
    case "dis":
        err := Display(dat, dis.RasArgs.DisArgs)
        
        if err != nil {
            return Handle(err, "failed to execute display")
        }
    
    case "ras":
        err := Raster(dat,  dis.RasArgs, dis.Sec)
        
        if err != nil {
            return Handle(err, "failed to execute raster")
        }
    }
    return nil
}

type(
    Coder struct {
        infile   string `pos:"0"`
        outfile  string `name:"out" default:""`
        metafile string `name:"meta" default:"geocode.json"`
        mode     string `name:"mode" default:""`
        GeoMeta
        CodeOpt
    }
)

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
*/

var PlotCmdFiles = map[string]Slice{
    "pwr": Slice{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": Slice{"slc", "rslc"},
    "mph": Slice{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": Slice{"hgt", "rdc"},
}

const (
    KeyErr Werror = "key '%s' is not present in '%s'"
    TypeErr Werror = "unexpected type %T for '%s', expected %s"
)

func (m JSONMap) String(name string) (ret string, err error) {
    tmp, ok := m[name]
    
    if !ok {
        err = KeyErr.Make(name, m)
        //err = Handle(nil, "key '%s' is not present in map '%s'", name, m)
        return
    }
    
    ret, ok = tmp.(string)
    
    if !ok {
        err = TypeErr.Make(tmp, name, "string")
        //err = Handle(nil, "unexpected type %T for '%s', expected string",
            //tmp, name)
        return
    }
    
    return ret, nil
}

func (m JSONMap) Int(name string) (ret int, err error) {
    tmp, ok := m[name]
    
    if !ok {
        err = KeyErr.Wrap(err, name, m)
        return
    }
    
    switch v := tmp.(type) {
    case int:
        return int(v), nil
    case int8:
        return int(v), nil
    case int16:
        return int(v), nil
    case int32:
        return int(v), nil
    case int64:
        return int(v), nil
    case float32:
        return int(v), nil
    case float64:
        return int(v), nil
    default:
        err = fmt.Errorf("failed to convert '%s' of type '%T' to int", tmp, tmp)
        return
    }
}
