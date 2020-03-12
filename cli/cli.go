package cli

import (
    "fmt"
    
    "github.com/bozso/gamma/sentinel1"
)

const (
    ParseError utils.CWerror = "failed to parse command line arguments"
    ParseErr utils.CWerror = "failed to parse command line arguments"
)


func SetupGammaCli(c *utils.Cli) {
    c.AddAction("like",
        "Initialize Gamma datafile with given datatype and shape.",
        &like{})
    
    c.AddAction("move",
        "Move a datafile and metadatafile to a given directory.",
        &move{})

    c.AddAction("make",
        "Create metafile for an existing datafile.",
        &create{})

    c.AddAction("coreg",
        "Coregister two Sentinel-1 SAR images.",
        &coreg{})    

    c.AddAction("select",
        "Select Sentinel-1 SAR zipfiles for processing.",
        &dataSelect{})    
}

type MetaFile struct {
    Meta string
}

func (m *MetaFile) SetCli(c *utils.Cli) {
    c.StringVar(&m.Meta, "meta", "", "Metadata json file")
}

type coreg struct {
    Master, Slave, Ref string 
    sentinel1.S1CoregOpt
}

func (co *coreg) SetCli(c *Cli) {
    co.S1CoregOpt.SetCli(c)
    
    c.StringVar(&co.Master, "master", "", "Master image.")
    c.StringVar(&co.Slave, "slave", "", "Slave image.")
    c.StringVar(&co.Ref, "ref", "", "Reference image.")
}

func (c coreg) Run() (err error) {
    var ferr = merr.Make("coregFn")
    
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
    
    if err = SaveJson("", &out.Ifg); err != nil {
        return ferr.Wrap(err)
    }
    
    fmt.Printf("Created RSLC: %s", out.Rslc.Tab)
    fmt.Printf("Created Interferogram: %s", out.Ifg.jsonName())
    
    return nil
}

/*

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
    var ferr = merr.Make("splitIfgFn")
    
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

*/

type Plotter struct {
    RasArgs
    Infile string `pos:"0"`
    PlotMode string `name:"mode"`
}

/*
func raster(args Args) (err error) {
    var ferr = merr.Make("raster")
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
*/

var PlotCmdFiles = map[string]Slice{
    "pwr": Slice{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": Slice{"slc", "rslc"},
    "mph": Slice{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": Slice{"hgt", "rdc"},
}
