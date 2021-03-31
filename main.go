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

    c.AddAction("rpc", "starts JSON RPC service", &gcli.JsonRPC{})

    //c.SetupGammaCli(cli)

    return c.Run()
}

func main() {
    cli.Run(Main{})
}
