package mli

import (
	"github.com/bozso/gomma/common"
	"github.com/bozso/gomma/data"
	"github.com/bozso/gomma/plot"
)

type MLI struct {
	data.FileWithPar `json:"MLI"`
}

func (m MLI) Validate() (err error) {
	return m.EnsureFloat()
}

// TODO: add loff, nlines
type Options struct {
	//Subset
	RefTab     string
	Looks      common.RngAzi
	WindowFlag bool
	plot.ScaleExp
}

/*
func (opt *Options) Parse() {
    opt.ScaleExp.Parse()

    if len(opt.RefTab) == 0 {
        opt.RefTab = "-"
    }

    opt.Looks.Default()
}
*/

func (_ MLI) PlotMode() (m plot.Mode) {
	return plot.Power
}
