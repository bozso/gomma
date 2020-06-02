package main

import (
    "fmt"
    "os"

    "github.com/bozso/gomma/cli"
)

func Main() error {
    c := cli.New("gamma",
        "Wrapper program for the GAMMA SAR processing software.")
        
    c.SetupGammaCli(cli)
    return c.Run(os.Args[1:])
}

func main() {
    cli.Run(Main)
}
