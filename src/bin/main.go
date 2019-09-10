package main

import (
	gm "../gamma"
	fl "flag"
	"fmt"
	"os"
)


func main() {
	defer gm.RemoveTmp()

	proc := fl.NewFlagSet("proc", fl.ExitOnError)

	conf := gm.NewConfig(proc)

	init := fl.NewFlagSet("init", fl.ExitOnError)
	path := init.String("config", "gamma.json",
		"Processing configuration file")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'proc' or 'init' subcommands!")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "proc":
		proc.Parse(os.Args[2:])
		procConf, start, stop, err := conf.Parse()

		gm.Fatal(err, "Could not parse configuration!")

		err = procConf.RunSteps(start, stop)
		gm.Fatal(err, "Error occurred while running processing steps!")

	case "init":
		init.Parse(os.Args[2:])

		err := gm.MakeDefaultConfig(*path)
		gm.Fatal(err, "Could not create config file: '%s'!", *path)

	default:
		fmt.Println("Expected 'proc' or 'init' subcommands!")
		os.Exit(1)
	}

	fmt.Println(gm.First())
}
