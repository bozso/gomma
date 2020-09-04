package cli

import (
    //"fmt"
    //gerrors "errors"
    
    "github.com/bozso/gotoolbox/errors"
    "github.com/bozso/gotoolbox/cli"
    //"github.com/bozso/gotoolbox/path"

    //s1 "github.com/bozso/gomma/sentinel1"
    "github.com/bozso/gomma/plot"
    //"github.com/bozso/gomma/date"
    //"github.com/bozso/gomma/common"
)

const (
    ParseError errors.String = "failed to parse command line arguments"
    ParseErr errors.String = "failed to parse command line arguments"
)


func SetupGammaCli(c *cli.Cli) {
    /*
    c.AddAction("coreg",
        "Coregister two Sentinel-1 SAR images.",
        &coreg{})

    c.AddAction("select",
        "Select Sentinel-1 SAR zipfiles for processing.",
        &dataSelect{})    
    */
}

type MetaFile struct {
    Meta string
}

func (m *MetaFile) SetCli(c *cli.Cli) {
    c.StringVar(&m.Meta, "meta", "", "Metadata json file")
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
    plot.RasArgs
    Infile string
    PlotMode string
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

var PlotCmdFiles = map[string][]string{
    "pwr": []string{"pix_sigma0", "pix_gamma0", "sbi_pwr", "cc", "rmli", "mli"},
    "SLC": []string{"slc", "rslc"},
    "mph": []string{"sbi", "sm", "diff", "lookup", "lt"},
    "hgt": []string{"hgt", "rdc"},
}
