package batch

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/cli"
)

type Operation interface {
    Call(infile path.ValidFile) (outfile path.ValidFile, err error)
}

type Creator interface {
    CreateOp(Context) (Operation, error)
}

type CreatorMap map[string]Creator

type Controller struct {
    creators CreatorMap
    ctx Context
    data []byte
}

func (c *Controller) SetCli(cl *cli.Cli) {
    cl.NewFlag().
        Name("profile").
        Usage("profile file that contains settings").
        Var(&c.ctx)

}
