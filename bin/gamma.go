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

    if len(os.Args) < 2 {
        fmt.Printf("Expected on of the following subcommands: %v!\n", commands)
        os.Exit(1)
    }
    
    mode := os.Args[1]
    
    switch mode {
    case "proc":
        proc, err := gm.NewProcess(os.Args[2:])
        
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
        init, err := gm.InitParse(os.Args[2:])
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
        batch, err := gm.NewBatcher(os.Args[2:])
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
        dis, err := gm.NewDisplayer(os.Args[1:])
        
        if err != nil {
            fmt.Printf("failed to parse plot arguments: %s!\n", err)
            return
        }
        
        err = dis.Plot()
        
        if err != nil {
            fmt.Printf("plotting failed: %s!\n", err)
            return
        }
    default:
        fmt.Printf("Expected on of the following subcommands: %v!\n", commands)
        return
    }

    return

}
