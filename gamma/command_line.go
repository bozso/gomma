package gamma

import (
    //"log"
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
        //"move": move,
        //"make": create,
        //"stat": stat,
        "splitIfg": splitIfg,
        "geocode": geocode,
        "raster": raster,
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
    StepErr Werror = "step '%s' not in list of available steps"
)


type UnrecognizedMode struct {
    name, got string
    Err error
}

func (e UnrecognizedMode) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got, e.name)
}

func (e UnrecognizedMode) Unwrap() error {
    return e.Err
}

type ModeError struct {
    name string
    got fmt.Stringer
    Err error
}

func (e ModeError) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got.String(), e.name)
}

func (e ModeError) Unwrap() error {
    return e.Err
}

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
            return
        }

        if istop == -1 {
            listSteps()
            err = StepErr.Make(proc.Stop)
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
        return
    }

    //fmt.Printf("%#v\n", batch)
    
    switch batch.Mode {
    case "quicklook":
        if err = batch.Quicklook(); err != nil {
            return
        }
    case "mli", "MLI":
        if err = batch.MLI(); err != nil {
            return
        }
    case "ras", "raster", "plot":
        if err = batch.Raster(); err != nil {
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
    
    var dat DatParFile
    if err = Load(path, &dat); err != nil {
        err = Handle(err, "failed to parse json metadatafile '%s'", path)
        return
    }
    
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

//type (
    //Coregister struct {
        //CoregOpt
        //Master  string `name:"master" default:""`
        //Slave   string `name:"slave" default:""`
        //Ref     string `name:"ref" default:""`
        //Outfile string `name:"out" default:""`
    //}

//)

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


//type Creater struct {
    //Dtype      string `name:"dtype"`
    //Ftype      string `name:"ftype"`
    //ParfileExt string `name:"parExt"`
    //DatParFile
    //MetaFile
//}

//func create(args Args) (err error) {
    //c := Creater{}
    
    //if err = args.ParseStruct(&c); err != nil {
        //err = ParseErr.Wrap(err)
        //return
    //}
    
    //dat := c.Dat
    
    //if len(dat) == 0 {
        //if err = c.DatParFile.Tmp(); err != nil {
            //return
        //}
    //}
    
    //ext := c.ParfileExt
    //if len(ext) == 0 {
        //ext = "par"
    //}
    
    //par := c.Par
    //if len(par) == 0 {
        //par = fmt.Sprintf("%s.%s", dat, ext)
    //}
    
    
    //dt_, dt := c.Dtype, Unknown
    
    //if len(dt_) > 0 {
        //dt, err = str2dtype(dt_)
        //if err != nil {
            //return
        //}
    //}
    
    //dat, err = fp.Abs(dat)
    //if err != nil { return }
    
    //par, err = fp.Abs(par)
    //if err != nil { return }
        
    //datf, err := NewDataFile(dat, par, dt) 
    //if err != nil {
        //err = DataCreateErr.Wrap(err, "DataFile")
        //return
    //}
    
    
    //var Dat DataFile
    //ftype := str.ToUpper(c.Ftype)
    //switch ftype {
    //case "MLI":
        //Dat = MLI{dataFile: datf}
    //default:
        //err = fmt.Errorf("unrecognized filetype '%s'", ftype)
        //return
    //}
    
    //if err = Dat.Save(c.Meta); err != nil {
        //err = Handle(err, "failed to save datafile to metafile '%s'",
            //c.MetaFile)
        //return
    //}
    
    //return nil
//}


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

func splitIfg(args Args) (err error) {
    si := SplitIfg{}
    si.Mode = Ifg
    
    if err := args.ParseStruct(&si); err != nil {
        return ParseErr.Wrap(err)
    }
    
    ms, ss := si.Master, si.Slave
    
    if len(ms) == 0 || len(ss) == 0 {
        return fmt.Errorf("both master and slave SLC files should be specified")
    }
    
    var m, s SLC
    
    if err = Load(ms, &m); err != nil {
        return
    }
    
    if err = Load(ss, &s); err != nil {
        return
    }
    
    id := ID(m, s, DShort)
    mode := str.ToUpper(si.SpectrumMode)
    
    switch mode {
    case "BEAM", "B":
        if err = SameShape(m, s); err != nil {
            return
        }
        
        out, err := m.SplitBeamIfg(s, si.SBIOpt)
        
        if err != nil {
            return err
        }
        
        if err = Save(id + "_sbi_mli.json", &out.Mli); err != nil {
            return err
        }
        
        if err = Save(id + "_sbi_ifg.json", &out.Ifg); err != nil {
            return err
        }
    //case "SPECTRUM", "S":
        //opt := si.SSIOpt
        
        //Mli, err := LoadDataFile(si.Mli)
        //if err != nil {
            //return err
        //}
        
        //mli, ok := Mli.(MLI)
        
        //if !ok {
            //return TypeErr.Make(Mli, "mli", "MLI")
        //}
        
        //out, err := m.SplitSpectrumIfg(s, mli, opt)
        
        //if err != nil {
            //return err
        //}
        
        // still need to figure out the returned files
        //return nil
    default:
        return fmt.Errorf("unrecognized Split Interferogram mode: '%s'", mode)
    }
    return nil
}

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


type(
    Coder struct {
        Lookup   string `pos:"0"`
        Infile   string `pos:"1"`
        Outfile  string `name:"out"`
        Mode     string `name:"mode"`
        Intpol   string `name:"intpol" default:"nearest"`
        Shape    string `name:"shape"`
        CodeOpt
    }
)

func geocode(args Args) (err error) {
    c := Coder{}
    
    if err = args.ParseStruct(&c); err != nil {
        err = ParseErr.Wrap(err)
        return
    }
    
    shape := c.Shape
    
    if len(shape) > 0 {
        var dat DatFile
        if err = Load(shape, &dat); err != nil {
            return
        }
        
        c.Rng = dat.Rng()
        c.Azi = dat.Azi()
    }
    
    imode := NearestNeighbour
    
    switch c.Intpol {
    case "nearest":
        imode = NearestNeighbour
    case "bic":
        imode = BicubicSpline
    case "bic_log":
        imode = BicubicSplineLog
    case "bic_sqrt":
        imode = BicubicSplineSqrt
    case "bsp":
        imode = BSpline
    case "bsp_sqrt":
        imode = BSplineSqrt
    case "lanc":
        imode = Lanczos
    case "lanc_sqrt":
        imode = LanczosSqrt
    case "inv_dist":
        imode = InvDist
    case "inv_sqrd_dist":
        imode = InvSquaredDist
    case "const":
        imode = Constant
    case "gauss":
        imode = Gauss
    default:
        err = UnrecognizedMode{name: "interpolation option", got: c.Intpol}
        return
    }
    
    c.InterpolMode = imode
    
    var l Lookup
    if err = Load(c.Lookup, &l); err != nil {
        return
    }
    
    var dat DatFile
    if err = Load(c.Infile, &dat); err != nil {
        return
    }
    
    mode := str.ToUpper(c.Mode)
    
    var out DatFile
    switch mode {
    case "TORADAR", "RADAR":
        if out, err = l.geo2radar(dat, c.CodeOpt); err != nil {
            return
        }
    case "TOGEO", "GEO":
        if out, err = l.radar2geo(dat, c.CodeOpt); err != nil {
            return
        }
    default:
        err = UnrecognizedMode{name: "geocoding", got: mode}
        return
    }
    
    if out, err = out.Move("."); err != nil {
        return
    }
    
    return Save(c.Outfile, &out)
}

type Plotter struct {
    RasArgs
    Infile string `pos:"0"`
    PlotMode string `name:"mode"`
}

func raster(args Args) (err error) {
    p := Plotter{}
    
    if err = args.ParseStruct(&p); err != nil {
        return ParseErr.Wrap(err)
    }
        
    mode, m := Undefined, str.ToUpper(p.PlotMode)
    
    switch m {
    case "PWR", "POWER":
        mode = Power
    case "MPH", "MAGPHASE":
        mode = MagPhase
    case "MPHPWR", "MAGPHASEPWR":
        mode = MagPhasePwr
    case "SLC", "SINGLELOOK":
        mode = SingleLook
    case "DB", "DECIBEL":
        mode = Decibel
    case "BYTE", "UCHAR":
        mode = Byte
    case "CC", "COHERENCE":
        mode = CC
    case "DT", "DEFORM":
        mode = Deform
    case "LIN", "LINEAR":
        mode = Linear
    case "HGT", "HEIGHT":
        mode = Height
    case "UNW", "UNWRAPPED":
        mode = Unwrapped
    }
    
    p.Mode = mode
    
    var dat DatFile
    if err = Load(p.Infile, &dat); err != nil {
        return
    }
    
    
    return dat.Raster(p.RasArgs)
}

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
