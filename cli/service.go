package cli

import (
    "github.com/bozso/gotoolbox/cli"

    "github.com/bozso/gomma/service"
)

type JsonRPC struct {
    jsonRpc cli.JsonRPC
}

func (j *JsonRPC) SetCli(c *cli.Cli) {
    j.jsonRpc.SetCli(c)
    
    j.jsonRpc.Add(service.DataSelect{})
}

func (j JsonRPC) Run() (err error) {
    return j.jsonRpc.Run()
}

