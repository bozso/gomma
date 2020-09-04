package main

import (
    //"fmt"
    //"os"

    gcli "github.com/bozso/gomma/cli"
    "github.com/bozso/gotoolbox/cli"
)

type Main struct{}

func (_ Main) Run() (err error) {
    c := cli.New("gamma",
        "Wrapper program for the GAMMA SAR processing software.")
    
    c.Add(&gcli.RPC{})
    
    //c.SetupGammaCli(cli)

    return c.Run(os.Args[1:])
}

func main() {
    cli.Run(Main{})
}
