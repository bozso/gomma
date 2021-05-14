package server

import (
	"fmt"
	"github.com/bozso/gotoolbox/path"

	ifg "github.com/bozso/gomma/interferogram"
	"github.com/bozso/gomma/mli"
	s1 "github.com/bozso/gomma/sentinel1"
	"github.com/bozso/gomma/slc"
)

type ConversionFailiures int

const (
	SLC ConversionFailiures = iota
	MLI
	S1SLC
	maxConvertFails
)

var ConvertFails = [maxConvertFails]ConvertFail{
	ConvertFail("SLC"),
	ConvertFail("MLI"),
}

type Unknown struct{}

func (_ Unknown) AsSLC() (s slc.SLC, err error) {
	err = ConvertFails[SLC]
	return
}
