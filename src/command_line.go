package gamma

import (
    //"log"
    //"reflect"
    "fmt"
    //"os"
    "strings"
    //"path/filepath"
    "github.com/mkideal/cli"
)

type (
    Cmd func(Args) error
    JSONMap map[string]interface{}
    
    MetaFile struct {
        Meta string `cli:"*meta" usage:"Metadata json file"`
    }
)

const (
    ParseErr CWerror = "failed to parse command line arguments"
)

var (
    BatchModes = []string{"quicklook", "mli / MLI", "ras"}
)

type (
    root struct {
        cli.Helper
    }
)

var Root = &cli.Command{
    Desc: "Gamma Golang wrapper command line program.",
    Argv: func() interface{} { return &root{} },
    Fn: func(ctx *cli.Context) error { return nil },
}

type process struct {
    cli.Helper
    Conf        string `cli:"conf" usage:"Processing configuration file" dft:"gamma.json"`
    Step        string `cli:"step" usage:"Execute single step"`
    Start       string `cli:"start" usage:"Starting step"`
    Stop        string `cli:"stop" usage:"Stoppong step"` 
    Log         string `cli:"log" usage:"Logfile" dft:"gamma.log"`
    CachePath   string `cli:"cache"`
    Skip        bool   `cli:"skip" usage: "Skip optional steps"`
    Show        bool   `cli:"show" usage:"Show available steps"`
    InputFile   string `cli:"in" usage:"Input file list"`
}

//var Process = &cli.Command{
    //Name: "proc",
    //Desc: "Execute processing steps",
    //Argv: func() interface{} { return &process{} },
    //Fn: proc,
//}

//func proc(ctx *cli.Context) (err error) {
    //proc := ctx.Argv().(*process) 
    
    //start, stop, err := proc.Parse()
    //if err != nil {
        //err = Handle(err, "failed to  parse processing steps")
        //return
    //}
    
    //if err = proc.RunSteps(start, stop); err != nil {
        //err = Handle(err, "error occurred while running processing steps")
        //return
    //}
    //return nil
//}

//func stepIndex(step string) int {
    //for ii, _step := range stepList {
        //if step == _step {
            //return ii
        //}
    //}
    //return -1
//}

//func listSteps() {
    //fmt.Println("Available processing steps: ", stepList)
//}


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

//func (proc process) Parse() (istart int, istop int, err error) {
    //if proc.Show {
        //listSteps()
        //os.Exit(0)
    //}

    //istep, istart, istop := stepIndex(proc.Step), stepIndex(proc.Start),
        //stepIndex(proc.Stop)

    //if istep == -1 {
        //if istart == -1 {
            //listSteps()
            //err = StepErr.Make(proc.Start)
            //return
        //}

        //if istop == -1 {
            //listSteps()
            //err = StepErr.Make(proc.Stop)
            //return
        //}
    //} else {
        //istart = istep
        //istop = istep + 1
    //}

    //return istart, istop, nil
//}

//func (proc process) RunSteps(start, stop int) (err error) {
    //config := Config{InFile: proc.InputFile}
    
    //if err = LoadJson(proc.Conf, &config); err != nil {
        //return
    //}
    
    //for ii := start; ii < stop; ii++ {
        //name := stepList[ii]
        //step := steps[name]

        //delim(fmt.Sprintf("START: %s", name), "*")

        //if err = step(&config); err != nil {
            //return Handle(err, "error while running step '%s'",
                //name)
        //}

        //delim(fmt.Sprintf("END: %s", name), "*")
    //}
    //return nil
//}


type initGamma struct {
    Outfile string `cli:"*out" usage:"Outfile"`
}

var Init = &cli.Command{
    Name: "init",
    Desc: "Initialize Gamma (pre)processing",
    Argv: func() interface{} { return &initGamma{} },
    Fn: InitGamma,
}

func InitGamma(ctx *cli.Context) (err error) {
    conf := ctx.Argv().(*initGamma).Outfile
    
    if err = MakeDefaultConfig(conf); err != nil {
        err = Handle(err, "failed to create config file")
    }
    return
}

//type(
    //Batcher struct {
        //Mode     string `cli:"*m,mode"`
        //Infile   string `cli:"*i,in"`
        //Conf     string `cli:"conf" dft:"gamma.conf"`
        //OutDir   string `cli:"out" dft:"."`
        //Filetype string `cli:"f,ftype"`
        //Mli      string `cli:"mli"`
        //RasArgs
        //Config
    //}
//)

//func batch(args Args) (err error) {
    //batch := Batcher{}
    
    //if err = args.ParseStruct(&batch); err != nil {
        //err = ParseErr.Wrap(err)
        //return
    //}

    ////fmt.Printf("%#v\n", batch)
    
    //switch batch.Mode {
    //case "quicklook":
        //if err = batch.Quicklook(); err != nil {
            //return
        //}
    //case "mli", "MLI":
        //if err = batch.MLI(); err != nil {
            //return
        //}
    //case "ras", "raster", "plot":
        //if err = batch.Raster(); err != nil {
            //return
        //}
    //default:
        //err = Handle(nil, "unrecognized mode: '%s'! Choose from: %v",
            //batch.Mode, BatchModes)
        //return
    //}
    //return nil
//}

//func (b Batcher) Quicklook() error {
    //cache := filepath.Join(b.General.CachePath, "sentinel1")

    //info := &ExtractOpt{root: cache, pol: b.General.Pol}

    //path := b.Infile
    //file, err := NewReader(path)

    //if err != nil {
        //return err
    //}

    //defer file.Close()

    //for file.Scan() {
        //line := file.Text()

        //s1, err := NewS1Zip(line, cache)

        //if err != nil {
            //return Handle(err, "failed to parse Sentinel-1 '%s'", s1.Path)
        //}

        //image, err := s1.Quicklook(info)

        //if err != nil {
            //return Handle(err, "failed to retreive quicklook file '%s'",
                //s1.Path)
        //}

        //fmt.Println(image)
    //}

    //return nil
//}

//func (b Batcher) MLI() (err error) {
    //path := b.InFile
    
    //var file FileReader
    //if file, err = NewReader(path); err != nil {
        //return
    //}

    //defer file.Close()
    
    //opt := &MLIOpt {
        //Looks: b.General.Looks,
    //}
    
    //mliDir := b.OutDir
    
    //if err = os.MkdirAll(mliDir, os.ModePerm); err != nil {
        //return Handle(err, "failed to create directory '%s'", mliDir)
    //}
    
    //for file.Scan() {
        //line := file.Text()
        
        
        //var s1 S1SLC
        //if s1, err = FromTabfile(line); err != nil {
            //err = Handle(err, "failed to parse S1SLC tabfile '%s'", line)
            //return
        //}
        
        //var name string
        //if name, err = filepath.Abs(filepath.Join(mliDir, s1.Format(DateShort))); err != nil {
            //return
        //}
        
        //var mli MLI
        //if mli, err = NewMLI(name + ".mli", ""); err != nil {
            //err = Handle(err, "could not create MLI struct")
            //return
        //}
        
        //var exist bool
        //if exist, err = mli.Exist(); err != nil {
            //return
        //}
        
        //json := name + ".json"
        
        //if exist {
            //if err = Save(json, &mli); err != nil {
                //return
            //}
            //continue
        //}
        
        //if err = s1.MLI(&mli, opt); err != nil {
            //return Handle(err, "failed to retreive MLI file for '%s'",
                //line)
        //}
        
        //if err = Save(json, &mli); err != nil {
            //return
        //}
        //fmt.Println(json) 
    //}

    //return nil
//}

//func (b Batcher) Raster() (err error) {
    //path := b.Infile
    //opt := b.RasArgs
    
    //var file FileReader
    //if file, err = NewReader(path); err != nil {
        //return err
    //}
    //defer file.Close()
    
    //for file.Scan() {
        //line := file.Text()
        
        //if len(line) == 0 {
            //continue
        //}
        
        //var data DatFile
        //if err = Load(line, &data); err != nil {
            //err = Handle(err, "failed to load datafile from '%s'", line)
            //return
        //}
        
        //if err = data.Raster(opt); err != nil {
            //err = Handle(err, "failed to create raster file for '%s'",
                //line)
            //return
        //}
    //}
    
    //return nil
//}

//type Like struct {
    //In    string `name:"in"`
    //Out   string `name:"out"`
    //Dtype string `name:"dtype"`
    //Ext   string `name:"ext" default:"dat"`
//}

//func like(args Args) (err error) {
    //l := Like{}
    
    //if err = args.ParseStruct(&l); err != nil {
        //err = ParseErr.Wrap(err)
        //return
    //}
    
    //in, out := l.In, l.Out
    
    //if len(in) == 0 && len(out) == 0 {
        //err = fmt.Errorf("expected parameter 'in' and 'out' to be set")
        //return
    //}
    
    //dt, dtype := l.Dtype, Unknown
    
    //if len(dt) > 0 {
        //dtype = str2dtype(dt)
    //}
    
    //var indat DatFile
    //if err = Load(in, &indat); err != nil {
        //return
    //}
    
    //if dtype == Unknown {
        //dtype = indat.Dtype()
    //}
    
    //if out, err = filepath.Abs(out); err != nil {
        //return
    //}
    
    //outdat := DatFile{
        //Dat: fmt.Sprintf("%s.%s", out, l.Ext),
        //URngAzi: indat.URngAzi,
        //DType: dtype,
    //}
    
    //return Save(out + ".json", &outdat)
//}


//type(
    //Mover struct {
        //OutDir   string `cli:"out" usage:"Output directory" dft:"."`
        //MetaFile
    //}
//)

//func move(args Args) (err error) {
    //m := Mover{}
    
    //if err = args.ParseStruct(&m); err != nil {
        //err = ParseErr.Wrap(err)
        //return
    //}
    
    //path := m.Meta
    
    //var dat DatParFile
    //if err = Load(path, &dat); err != nil {
        //err = Handle(err, "failed to parse json metadatafile '%s'", path)
        //return
    //}
    
    //out := m.OutDir
    
    //if dat, err = dat.Move(out); err != nil {
        //return err
    //}
    
    //if path, err = Move(path, out); err != nil {
        //return err
    //}
    
    //if err = SaveJson(path, dat); err != nil {
        //err = Handle(err, "failed to refresh json metafile")
        //return err
    //}
    
    //return nil
//}

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
    mode := strings.ToUpper(si.SpectrumMode)
    
    switch mode {
    case "BEAM", "B":
        if err = SameShape(m, s); err != nil {
            return
        }
        
        var out SBIOut
        if out, err = m.SplitBeamIfg(s, si.SBIOpt); err != nil {
            return err
        }
        
        if err = Save(id + "_sbi_mli.json", &out.Mli); err != nil {
            return
        }
        
        if err = Save(id + "_sbi_ifg.json", &out.Ifg); err != nil {
            return
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
    
    mode := strings.ToUpper(c.Mode)
    
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
        
    mode, m := Undefined, strings.ToUpper(p.PlotMode)
    
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
        return
    }
    
    ret, ok = tmp.(string)
    
    if !ok {
        err = TypeErr.Make(tmp, name, "string")
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
