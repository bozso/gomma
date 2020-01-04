package main

import (
    "fmt"
    "os"

    gamma "../src"
)

func main() {
    cli := gamma.NewCli("gamma",
        "Wrapper program for the GAMMA SAR processing software.")
    cli.SetupGammaCli()
    
    if err := cli.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error occurred in main while running: %s\n",
            err)
        os.Exit(1)
    }
}
