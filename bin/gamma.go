package main

import (
    "fmt"
    "log"
    "os"
    gm "../gamma"
    ref "reflect"
)

var commands = []string{"proc", "list", "init", "batch", "ras", "dis", "iono"}

func pp(ptr interface{}) error {
    vv := ref.ValueOf(ptr)
    kind := vv.Kind()
    
    if kind != ref.Ptr {
        return gm.Handle(nil, "expected a pointer to struct not '%v'", kind)
    }
    
    v := vv.Elem()
    t := v.Type()
    
    
    for ii := 0; ii < v.NumField(); ii++ {
        field := t.Field(ii)
        tag := field.Tag
        
        fmt.Printf("Tag: %s\n", tag)
    }
    
    return nil
}

const (
    parseErr = "failed to parse command line arguments: %s\n"
)

func main() {
    defer gm.RemoveTmp()
    
    if len(os.Args) < 2 {
        fmt.Printf("Expected on of the following subcommands: %v!\n", commands)
        return
    }
    
    mode := os.Args[1]    
    args := gm.NewArgs(os.Args[2:])
    
    //args.ParseStruct(&gm.SLC{})
    
    
    switch mode {
    case "proc":
        proc := gm.Process{}
        
        args.ParseStruct(&proc)
        
        start, stop, err := proc.Parse()
        if err != nil {
            log.Printf("failed to  parse processing steps: %s\n", err)
            return
        }
        
        err = proc.RunSteps(start, stop)
        
        if err != nil {
            log.Printf("error occurred while running processing steps: %s\n",
                err)
            return
        }

    case "init":
        init, err := gm.InitParse(args[2:])
        if err != nil {
            log.Printf(parseErr, err)
            return
        }

        err = gm.MakeDefaultConfig(init)
        if err != nil {
            log.Printf("failed to create config file '%s'!: %s\n",
                init, err)
            return
        }

    case "batch":
        batch := gm.Batcher{}
        
        if err = args.ParseStruct(&bath); err != nil {
            log.Printf(parseErr, err)
            return
        }

        switch batch.Mode {
        case "quicklook":
            if err = batch.Quicklook(); err != nil {
                log.Printf("Error: %w\n", err)
                return
            }
        case "mli", "MLI":
            if err = batch.MLI(); err != nil {
                log.Printf("Error: %s\n", err)
                return
            }
        
        case "ras", "raster", "plot":
            if err = batch.Raster(); err != nil {
                log.Printf("Error: %s\n", err)
                return
            }
        default:
            log.Printf("unrecognized mode: '%s'! Choose from: %v", batch.Mode,
                gm.BatchModes)
            return
        }
    case "ras", "dis":
        dis, err := gm.NewDisplayer(args[1:])
        
        if err != nil {
            fmt.Printf("failed to parse plot arguments: %s!\n", err)
            return
        }
        
        err = dis.Plot()
        
        if err != nil {
            fmt.Printf("plotting failed: %s!\n", err)
            return
        }
    case "iono":
        err := iono()
        
        if err != nil {
            fmt.Printf("iono failed: %s!\n", err)
            return
        }
    default:
        fmt.Printf("Expected on of the following subcommands: %v!\n", commands)
        return
    }
}

type (
    S1Pair struct {
        master, slave gm.S1SLC
    }
    
    SLCPair struct {
        master, slave gm.SLC
    }
)

func iono() error {
    var (
        err error
        orig S1Pair
    )
    
    if orig.master, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    if orig.slave, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    var deramp S1Pair

    deramp.master, err = orig.master.DerampRef()
    if  err != nil {
        return gm.Handle(err, "failed to deramp master S1SLC")
    }
    
    deramp.slave, err = orig.slave.DerampSlave(&orig.master, gm.RngAzi{}, false)
    if err != nil {
        return gm.Handle(err, "failed to deramp slave S1SLC")
    }
    
    const (
        rslc, ifg, hgt = "RSLC", "IFG", "dem.dem"
    )
    
    if err = os.MkdirAll(rslc, os.ModePerm); err != nil {
        return gm.Handle(err, "failed to create directory '%s'", rslc)
    }
    
    if err = os.MkdirAll(ifg, os.ModePerm); err != nil {
        return gm.Handle(err, "failed to create directory '%s'", ifg)
    }
    
    mID, sID := orig.master.Format(gm.DateShort), orig.slave.Format(gm.DateShort)
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    
    coreg := gm.S1Coreg {
        Tab: deramp.master.Tab,
        ID: mID,
        OutDir: ".",
        RslcPath: rslc,
        IfgPath: ifg,
        Hgt: hgt,
        Poly1: "-",
        Poly2: "-",
        Looks: gm.RngAzi{Rng:1, Azi:1},
        Clean: false,
        CoregOpt: gm.CoregOpt{
            CoherenceThresh:  0.8,
            FractionThresh:   0.01,
            PhaseStdevThresh: 0.8,
        },
    }
    
    var out gm.CoregOut
    if out, err = coreg.Coreg(&deramp.slave, nil); err != nil {
        return gm.Handle(err, "coregistration failed")
    }
    
    if !out.Ok {
        return gm.Handle(err, "coregistration failed")
    }
    
    lookup := ID + ".lt_fine"
    
    var slc SLCPair
    
    mopts := gm.MosaicOpts{Looks: gm.RngAzi{}}
    
    if slc.master, err = deramp.master.Mosaic(mopts); err != nil {
        return gm.Handle(err, "failed to mosaic master S1SLC")
    }
    
    if slc.slave, err = deramp.slave.Mosaic(mopts); err != nil {
        return gm.Handle(err, "failed to mosaic slave S1SLC")
    }
    
    var mmli gm.MLI
    if mmli, err = slc.master.MakeMLI(gm.MLIOpt{}); err != nil {
        return gm.Handle(err, "failed to create master MLI")
    }
    
    ssiOpt := gm.SSIOpt{
        Hgt: hgt,
        LtFine: lookup,
        OutDir: ".",
        Mode: gm.IfgUnwrapped,
    }
    
    //ssiOut, err := slc.master.SplitSpectrumIfg(slc.slave, mmli, ssiOpt)
    _, err = slc.master.SplitSpectrumIfg(slc.slave, mmli, ssiOpt)
    
    if err != nil {
        return gm.Handle(err, "SSI_INT failed")
    }
    
    if err = ssiOut.ifg.Move("."); err != nil {
        return gm.Handle(err, "failed to move SSI IFG")
    }
    
    if err = ssiOut.unw.Move("."); err != nil {
        return gm.Handle(err, "failed to move SSI unwrapped IFG")
    }
    return nil
}
