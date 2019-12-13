package gamma

import (
    "errors"
    "fmt"
    "strings"
    "path/filepath"
    "github.com/mkideal/cli"

    //"os"
    //"log"
    //"reflect"
)

type (
    Cmd func(Args) error
    JSONMap map[string]interface{}
    
    MetaFile struct {
        Meta string `cli:"*meta" usage:"Metadata json file"`
    }
)

var ParseError = errors.New("failed to parse command line arguments")

const (
    ParseErr CWerror = "failed to parse command line arguments"
)

var (
    BatchModes = []string{"quicklook", "mli / MLI", "ras"}
)

type root struct {
    cli.Helper
}


var Root = &cli.Command{
    Desc: "Gamma Golang wrapper command line program.",
    Argv: func() interface{} { return &root{} },
    Fn: func(ctx *cli.Context) error { return nil },
}

type UnrecognizedMode struct {
    name, got string
    err error
}

func (e UnrecognizedMode) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got, e.name)
}

func (e UnrecognizedMode) Unwrap() error {
    return e.err
}

type ModeError struct {
    name string
    got fmt.Stringer
    err error
}

func (e ModeError) Error() string {
    return fmt.Sprintf("unrecognized mode '%s' for %s", e.got.String(), e.name)
}

func (e ModeError) Unwrap() error {
    return e.err
}

var Like = &cli.Command{
    Name: "like",
    Desc: "Initialize Gamma datafile with given datatype and shape",
    Argv: func() interface{} { return &like{} },
    Fn: likeFn,
}

type like struct {
    In    string `cli:"*i,in" usage:"Reference metadata file"`
    Out   string `cli:"*o,out" usage:"Output metadata file"`
    Dtype DType  `cli:"d,dtype" usage:"Output file datatype" dft:"Uknown"`
    Ext   string `cli:"e,ext" usage:"Extension of datafile" dft:"dat"`
}

func likeFn(ctx *cli.Context) (err error) {
    var ferr = merr("likeFn")
    
    l := ctx.Argv().(*like)
    
    in, out := l.In, l.Out
    
    var indat DatFile
    if err = Load(in, &indat); err != nil {
        return ferr.Wrap(err)
    }
    
    dtype := l.Dtype

    if dtype == Unknown {
        dtype = indat.Dtype()
    }
    
    if out, err = filepath.Abs(out); err != nil {
        return ferr.Wrap(err)
    }
    
    outdat := DatFile{
        Dat: fmt.Sprintf("%s.%s", out, l.Ext),
        URngAzi: indat.URngAzi,
        DType: dtype,
    }
    
    if err = Save(out, &outdat); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

var MoveFile = &cli.Command{
    Name: "move",
    Desc: "Move a datafile and metadatafile to a given directory",
    Argv: func() interface{} { return &move{} },
    Fn: moveFn,
}

type move struct {
    OutDir   string `cli:"out" usage:"Output directory" dft:"."`
    MetaFile
}


func moveFn(ctx *cli.Context) (err error) {
    var ferr = merr("moveFn")
    
    m := ctx.Argv().(*move)
    path := m.Meta
    
    var dat DatParFile
    if err = Load(path, &dat); err != nil {
        return ferr.WrapFmt(err,
            "failed to parse json metadatafile '%s'", path) 
    }
    
    out := m.OutDir
    
    if dat, err = dat.Move(out); err != nil {
        return ferr.Wrap(err)
    }
    
    if path, err = Move(path, out); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = SaveJson(path, dat); err != nil {
        return ferr.WrapFmt(err, "failed to refresh json metafile")
    }
    
    return nil
}


var Coreg = &cli.Command{
    Name: "coreg",
    Desc: "Coregister two Sentinel-1 SAR images",
    Argv: func() interface{} { return &coreg{} },
    Fn: coregFn,
}

type coreg struct {
    S1CoregOpt
    Master  string `cli:"*m,master"`
    Slave   string `cli:"*s,slave"`
    Ref     string `cli:"ref"`
}

func coregFn(ctx *cli.Context) (err error) {
    var ferr = merr("coregFn")
    
    c := ctx.Argv().(*coreg)
    
    sm, ss, sr := c.Master, c.Slave, c.Ref
    
    var ref *S1SLC
    
    if len(sr) == 0 {
        ref = nil
    } else {
        var ref_ S1SLC
        if ref_, err = FromTabfile(sr); err != nil {
            return ferr.Wrap(err)
        }
        ref = &ref_
    }
    
    var s, m S1SLC
    if s, err = FromTabfile(ss); err != nil {
        return ferr.Wrap(err)
    }
    
    if m, err = FromTabfile(sm); err != nil {
        return ferr.Wrap(err)
    }
    
    c.Tab, c.ID = m.Tab, m.Format(DateShort)
    
    var out S1CoregOut
    if out, err = c.Coreg(&s, ref); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = Save("", &out.Ifg); err != nil {
        return ferr.Wrap(err)
    }
    
    fmt.Printf("Created RSLC: %s", out.Rslc.Tab)
    fmt.Printf("Created Interferogram: %s", out.Ifg.jsonName())
    
    return nil
}

var Create = &cli.Command{
    Name: "make",
    Desc: "Create metafile for an existing datafile",
    Argv: func() interface{} { return &create{} },
    Fn: createFn,
}

type create struct {
    Dat        string `cli:"d,dat"`
    Par        string `cli:"p,par"`
    Dtype      DType  `cli:"D,dtype" dft:"Unknown"`
    Ftype      string `cli:"f,ftype"`
    Ext        string `cli:"P,parExt" dft:"par"`
    MetaFile
}

func createFn(ctx *cli.Context) (err error) {
    var ferr = merr("createFn")
    c := ctx.Argv().(*create)
    
    var dat string
    if dat, err = filepath.Abs(c.Dat); err != nil {
        return ferr.Wrap(err)
    }
    
    par := c.Par
    if len(par) > 0 {
        if par, err = filepath.Abs(par); err != nil {
            return ferr.Wrap(err)
        }
    }
    
    var datf DatParFile
    if datf, err = NewDatParFile(dat, par, c.Ext, c.Dtype); err != nil {
        err = StructCreateError.Wrap(err, "DatParFile")
        return ferr.Wrap(err)
    }
    
    if err = datf.Parse(); err != nil {
        return ferr.Wrap(err)
    }
    
    if datf.DType, err = datf.ParseDtype(); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = Save(c.Meta, &datf); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}


var SplitIFG = &cli.Command{
    Name: "splitIfg",
    Desc: "Split Beam/Spectrum Interferometry",
    Argv: func() interface{} { return &splitIfg{} },
    Fn: splitIfgFn,
}

type splitIfg struct {
    SBIOpt
    SSIOpt
    SpectrumMode string `cli:"M,mode"`
    Master       string `cli:"*m,master"`
    Slave        string `cli:"*s,slave"`
    Mli          string `cli:"mli"`
}


func splitIfgFn(ctx *cli.Context) (err error) {
    var ferr = merr("splitIfgFn")
    
    si := ctx.Argv().(*splitIfg)

    ms, ss := si.Master, si.Slave
    
    var m, s SLC

    if err = Load(ms, &m); err != nil {
        return ferr.Wrap(err)
    }

    if err = Load(ss, &s); err != nil {
        return ferr.Wrap(err)
    }
    
    mode := strings.ToUpper(si.SpectrumMode)
    
    switch mode {
    case "BEAM", "B":
        if err = SameShape(m, s); err != nil {
            return ferr.Wrap(err)
        }
        
        if err = m.SplitBeamIfg(s, si.SBIOpt); err != nil {
            return ferr.Wrap(err)
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
        err = UnrecognizedMode{name:"Split Interferogram", got: mode}
        return ferr.Wrap(err)
    }
    return nil
}

type (
    Stat struct {
        Out string `cli:"*o,out" usage:"Output file`
        Subset
        MetaFile
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


var GeoCode = &cli.Command{
    Name: "geocode",
    Desc: "",
    Argv: func() interface{} { return &geoCode{} },
    Fn: geoCodeFn,
}

type geoCode struct {
    Lookup   string `cli:"*l,lookup" usage:"Lookup table file"`
    Infile   string `cli:"*infile" usage:"Input datafile"`
    Outfile  string `cli:"*out" usage:"Output datafile"`
    Mode     string `cli:"mode" usage:"Geocode direction; from or to radar cordinates"`
    Shape    string `cli:"s,shape" usage:"Shape of the output file"`
    CodeOpt
}


func geoCodeFn(ctx *cli.Context) (err error) {
    var ferr = merr("geoCodeFn")
    
    c := ctx.Argv().(*geoCode)
    
    shape := c.Shape
    
    if len(shape) > 0 {
        var dat DatFile
        if err = Load(shape, &dat); err != nil {
            return
        }
        
        c.Rng = dat.Rng()
        c.Azi = dat.Azi()
    }
    
    var l Lookup
    if err = Load(c.Lookup, &l); err != nil {
        return ferr.Wrap(err)
    }
    
    var dat DatFile
    if err = Load(c.Infile, &dat); err != nil {
        return ferr.Wrap(err)
    }
    
    mode := strings.ToUpper(c.Mode)
    
    var out DatFile
    switch mode {
    case "TORADAR", "RADAR":
        if out, err = l.geo2radar(dat, c.CodeOpt); err != nil {
            return ferr.Wrap(err)
        }
    case "TOGEO", "GEO":
        if out, err = l.radar2geo(dat, c.CodeOpt); err != nil {
            return ferr.Wrap(err)
        }
    default:
        err = UnrecognizedMode{name: "geocoding", got: mode}
        return ferr.Wrap(err)
    }
    
    if out, err = out.Move("."); err != nil {
        return ferr.Wrap(err)
    }
    
    if err = Save(c.Outfile, &out); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

type Plotter struct {
    RasArgs
    Infile string `pos:"0"`
    PlotMode string `name:"mode"`
}

func raster(args Args) (err error) {
    var ferr = merr("raster")
    p := Plotter{}
    
    if err = args.ParseStruct(&p); err != nil {
        return ferr.Wrap(err)
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
        return ferr.Wrap(err)
    }
    
    if err = dat.Raster(p.RasArgs); err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
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
    var ferr = merr("JSONMap.String")
    tmp, ok := m[name]
    
    if !ok {
        err = ferr.Wrap(KeyErr.Make(name, m))
        return
    }
    
    ret, ok = tmp.(string)
    
    if !ok {
        err = ferr.Wrap(TypeErr.Make(tmp, name, "string"))
        return
    }
    
    return ret, nil
}

func (m JSONMap) Int(name string) (ret int, err error) {
    var ferr = merr("JSONMap.Int")
    
    tmp, ok := m[name]
    
    if !ok {
        err = ferr.Wrap(KeyErr.Wrap(err, name, m))
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
        err = ferr.WrapFmt(err,
            "failed to convert '%s' of type '%T' to int", tmp, tmp)
        return
    }
}
