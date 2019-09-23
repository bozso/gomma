package main

import (
    "fmt"
    "log"
    "os"
    gm "../gamma"
)

func main() {
    defer gm.RemoveTmp()

    if len(os.Args) < 2 {
        fmt.Println("Expected 'proc', 'list' or 'init' subcommands!")
        os.Exit(1)
    }
    mode := os.Args[1]
    
    if mode == "ras" || mode == "dis" {
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
        return
    }
    
    
    switch os.Args[1] {
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

    case "list":
        list, err := gm.NewLister(os.Args[2:])
        if err != nil {
            log.Printf("failed to parse command line arguments: %s\n",
                err)
            return
        }

        switch list.Mode {
        case "quicklook":
            err = list.Quicklook()
            
            if err != nil {
                log.Printf("Error: %w", err)
                return
            }
        default:
            log.Printf("unrecognized mode: '%s'! Choose from: %v", list.Mode,
                gm.ListModes)
            return
        }
        
        /*
                if err != nil {
                    return
                }
                log.Printf(err, "Could not create config file: '%s'!", *path)
        */
    default:
        fmt.Println("Expected 'proc', 'list' or 'init' subcommands!")
        return
    }

    return

}
