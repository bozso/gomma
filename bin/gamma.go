package main

import (
    "fmt"
    "log"
    "os"
    gm "../gamma"
)

var commands = []string{"proc", "list", "init", "batch", "ras", "dis"}

func main() {
    defer gm.RemoveTmp()
    
    args := os.Args
    
    if len(args) < 2 {
        fmt.Printf("Expected on of the following subcommands: %v!\n", commands)
        os.Exit(1)
    }
    
    mode := args[1]
    
    switch mode {
    case "proc":
        proc, err := gm.NewProcess(args[2:])
        
        if err != nil {
            log.Printf("failed to parse command line arguments: %s\n", err)
            return
        }
        
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
            log.Printf("failed to parse command line arguments: %s\n",
                err)
            return
        }

        err = gm.MakeDefaultConfig(init)
        if err != nil {
            log.Printf("failed tocreate config file '%s'!: %s\n",
                init, err)
            return
        }

    case "batch":
        batch, err := gm.NewBatcher(args[2:])
        if err != nil {
            log.Printf("failed to parse command line arguments: %s\n",
                err)
            return
        }

        switch batch.Mode {
        case "quicklook":
            err = batch.Quicklook()
            
            if err != nil {
                log.Printf("Error: %w\n", err)
                return
            }
        case "mli", "MLI":
            err = batch.MLI()
            
            if err != nil {
                log.Printf("Error: %s\n", err)
                return
            }
        
        case "ras", "raster", "plot":
            err = batch.Raster()
            
            if err != nil {
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

var ssi_int = gm.Gamma.Must("SSI_INT")

func iono() error {
    var (
        err error
        ref, slave gm.S1SLC
    )
    
    if ref, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    if slave, err = gm.FromTabfile(""); err != nil {
        return gm.Handle(err, "failed to import S1SLC struct")
    }
    
    var (
        rslave gm.S1SLC
        out gm.CoregOut
    )
    
    const (
        rslc, ifg, hgt = "RSLC", "IFG", "dem.dem"
    )
    
    err = os.MkdirAll(rslc, os.ModePerm)
    
    if err != nil {
        return gm.Handle(err, "failed to create directory '%s'", rslc)
    }
    
    err = os.MkdirAll(ifg, os.ModePerm)
    
    if err != nil {
        return gm.Handle(err, "failed to create directory '%s'", ifg)
    }
    
    mID, sID := ref.Format(gm.DateShort), slave.Format(gm.DateShort)
    ID := fmt.Sprintf("%s_%s", mID, sID)
    
    
    coreg := gm.S1Coreg {
        Tab: ref.Tab,
        ID: mID,
        OutDir: ".",
        RslcPath: rslc,
        IfgPath: ifg,
        Hgt: hgt,
        Poly1: "-",
        Poly2: "-",
        Looks: gm.RngAzi{Rng:1, Azi:1},
        Clean: true,
        CoregOpt: gm.CoregOpt{
            CoherenceThresh:  0.8,
            FractionThresh:   0.01,
            PhaseStdevThresh: 0.8,
        },
    }
    
    if out, err = coreg.Coreg(&slave, &ref); err != nil {
        return gm.Handle(err, "coregistration failed")
    }
    
    if !out.Ok {
        return gm.Handle(err, "coregistration failed")
    }
    
    dslave, dref := out.Rslc, gm.S1SLC{}
    
    lookup := ID + ".lt_fine"
    
    if dref, err = ref.DerampRef(); err != nil {
        return gm.Handle(err, "failed to deramp master S1SLC")
    }
    
    if dslave, err = rslave.DerampSlave(&ref, gm.RngAzi{}, false); err != nil {
        return gm.Handle(err, "failed to deramp slave S1SLC")
    }
    
    const IFGOnly, IFGAndUnwrapped = 1, 2
    
    _, err = ssi_int(rslc.Dat, rslc.Par, rmli.Dat, rmli.Par, hgt, lookup,
                     off, sslc.Dat, sslc.Par, IFGAndUnwrapped, mID, sID,
                     ID, ".", 1)
    
    if err != nil {
        return Handle(err, "SSI_INT failed")
    }
    
    return nil
}
