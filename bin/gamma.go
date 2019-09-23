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
        dis, err := NewDisplayer(os.Args[1:])
        
        if err != nil {
            fmt.Println("Failed to parse plot arguments!\n%s!", err)
            return
        }
        
        err = dis.Plot()
        
        if err != nil {
            fmt.Println("Plotting failed!\n%s!", err)
            return
        }
    }
    
    
	switch os.Args[1] {
	case "proc":
		proc, verr := gm.NewProcess(os.Args[2:])
        
        start, stop, err := proc.Parse()
    
		if err != nil {
			log.Printf("Error parsing processing steps!\nError: %s", err)
			return
		}
        
		err = proc.RunSteps(start, stop)
        
		if err != nil {
			log.Printf(
				"Error occurred while running processing steps!\nError: %s",
				err)
			return
		}

	case "init":
		init, err := gm.InitParse(os.Args[2:])
        if err != nil {
            log.Printf("Failed to parse command line arguments!\nError: %s",
                err)
            return
        }

		err = gm.MakeDefaultConfig(init)
		if err != nil {
			log.Printf("Could not create config file: '%s'!\nError: %s",
				init, err)
			return
		}

	case "list":
		list, err := gm.NewLister(os.Args[2:])
        if err != nil {
            log.Printf("Failed to parse command line arguments!\nError: %s",
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
            log.Printf("Unrecognized mode: %s! Choose from: %v", list.Mode,
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
