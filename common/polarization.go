package common

import (
	"github.com/bozso/gotoolbox/errors"

	"strings"
)

type Pol int

const (
	VV Pol = iota
	VH
	HV
	HH
	AllPolarisation
)

func (p *Pol) Set(s string) (err error) {

	const mode errors.Mode = "polarization"
	ps := strings.ToLower(s)

	switch ps {
	case "vv":
		*p = VV
	case "vh":
		*p = VH
	case "hv":
		*p = HV
	case "hh":
		*p = HH
	case "all":
		*p = AllPolarisation
	default:
		err = mode.Error(s)
	}
	return
}

func (p Pol) String() (s string) {
	switch p {
	case VV:
		s = "vv"
	case VH:
		s = "vh"
	case HV:
		s = "hv"
	case HH:
		s = "hh"
	case AllPolarisation:
		s = "-"
	}
	return
}
