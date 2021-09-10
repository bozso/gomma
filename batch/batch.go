package batch

import (
	"github.com/bozso/gotoolbox/cli"
)

type Operation interface {
	BatchOp(ctx Context, infile string) (outfile string, err error)
}

type OperationMap map[string]Operation

type Controller struct {
	operations OperationMap
	ctx        Context
	op         string
	infile     string
	outfile    string
	config     Config
}

func (c *Controller) SetCli(cl *cli.Cli) {
	cl.NewFlag().
		Name("profile").
		Usage("profile file that contains settings").
		Var(&c.ctx)

	cl.NewFlag().
		Name("operation").
		Usage("operation to carry out").
		String(&c.op)

	cl.NewFlag().
		Name("in").
		Usage("inputfile for batch operation").
		String(&c.infile, "-")

	cl.NewFlag().
		Name("out").
		Usage("outputfile of batch operation").
		StringVar(&c.infile, "-")
}
