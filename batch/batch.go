package batch

import (
	"fmt"
	"io"

	"github.com/bozso/gotoolbox/cli"
)

type Operation interface {
	BatchOp(ctx Context, in io.Reader, out io.Writer) (err error)
}

type OperationMap map[string]Operation

type Controller struct {
	operations OperationMap
	ctx        Context
	op         string
	infile     string
	outfile    string
}

func (c Controller) Run() (err error) {
	return WrapErr(c.op, func() (err error) {
		op, ok := c.operations[c.op]
		if !ok {
			return fmt.Errorf("operation '%s' not found", c.op)
		}



		err = op.BatchOp(c.ctx, )
		return
	})
}

func (c *Controller) SetCli(cl *cli.Cli) {
	cl.NewFlag().
		Name("profile").
		Usage("profile file that contains settings").
		Var(&c.ctx)

	cl.NewFlag().
		Name("operation").
		Usage("operation to carry out").
		StringVar(&c.op, "")

	cl.NewFlag().
		Name("in").
		Usage("inputfile for batch operation").
		StringVar(&c.infile, "-")

	cl.NewFlag().
		Name("out").
		Usage("outputfile of batch operation").
		StringVar(&c.infile, "-")
}
