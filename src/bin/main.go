package main

import (
	"fmt"
	"os"
    "log"
	gm "../gamma"
	fl "flag"
)


func main() {
	defer gm.RemoveTmp()

	proc := fl.NewFlagSet("proc", fl.ExitOnError)

	conf := gm.NewConfig(proc)

	init  := fl.NewFlagSet("init", fl.ExitOnError)
	cpath := init.String("config", "gamma.json", "Processing configuration file")
	
    quick  := fl.NewFlagSet("quicklook", fl.ExitOnError)
	mpath  := quick.String("meta", "meta.json", "Processing metadata file.")
    cache  := quick.String("cache", gm.DefaultCachePath, "Path to cached files.")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'proc' or 'init' subcommands!")
		os.Exit(1)
	}
    
    meta := gm.S1ProcData{}
    
	switch os.Args[1] {
	case "proc":
		proc.Parse(os.Args[2:])
		procConf, start, stop, err := conf.Parse()

		gm.Fatal(err, "Could not parse configuration!")

		err = procConf.RunSteps(start, stop)
		
        if err != nil {
            log.Printf(
                "Error occurred while running processing steps!\nError: %w",
                err)
            return
        }

	case "init":
		init.Parse(os.Args[2:])

		err := gm.MakeDefaultConfig(*cpath)
		if err != nil {
            log.Printf("Could not create config file: '%s'!\nError: %w",
                *cpath, err)
            return
        }
    
    case "quicklook":
		quick.Parse(os.Args[2:])
        
        err := gm.LoadJson(*mpath, &meta)
        if err != nil {
            log.Printf("Could not parse json file: '%s'!\nError: %w",
                *mpath, err)
            return
        }
        
        err = meta.Quicklook(*cache)
        if err != nil {
            log.Printf("Quicklook failed!\nError: %w", err)
            return
        }
        
        
        /*
        if err != nil {
            return 
        }
		log.Printf(err, "Could not create config file: '%s'!", *path)
        */
    
	default:
		fmt.Println("Expected 'proc', 'quicklook' or 'init' subcommands!")
		return
	}
    
    return
    
	fmt.Println(gm.First())
}
