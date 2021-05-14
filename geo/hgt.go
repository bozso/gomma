package geo

import (
	"github.com/bozso/gomma/data"
	"github.com/bozso/gomma/plot"
)

type Height struct {
	data.File
}

func (h Height) Validate() (err error) {
	return h.EnsureFloat()
}

func (h *Height) Set(s string) (err error) {
	return
}

func (_ Height) PlotMode() (m plot.Mode) {
	return plot.Height
}
