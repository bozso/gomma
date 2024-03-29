package server

import (
	"fmt"
	"github.com/bozso/gotoolbox/path"

	ifg "github.com/bozso/gomma/interferogram"
	"github.com/bozso/gomma/mli"
	s1 "github.com/bozso/gomma/sentinel1"
	"github.com/bozso/gomma/slc"
)

type S1SLC struct {
	Unknown
	slc s1.SLC
}

func (s S1SLC) AsS1SLC() (slc s1.SLC, err error) {
	return s.slc, nil
}
