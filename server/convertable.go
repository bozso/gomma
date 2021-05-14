package server

import (
	"fmt"
	"github.com/bozso/gotoolbox/path"

	ifg "github.com/bozso/gomma/interferogram"
	"github.com/bozso/gomma/mli"
	s1 "github.com/bozso/gomma/sentinel1"
	"github.com/bozso/gomma/slc"
)

type Convertable interface {
	AsSLC() (slc.SLC, error)
	AsMLI() (mli.MLI, error)
	AsS1SLC() (s1.SLC, error)
	AsInterferogram() (ifg.File, error)
}

type ConvertFail string

func (c ConvertFail) Error() (s string) {
	return fmt.Sprintf("failed to convert datafile to %s", string(c))
}
