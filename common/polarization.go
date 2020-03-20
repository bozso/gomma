package common

import (

)

type Pol int

const (
    VV Pol = iota
    VH
    HV
    HH
)

func (p *Pol) Set(s string) (err error) {
    ps := strings.ToLower(s)
    
    switch ps {
    case "vv":
        *p = VV
    case "vh":
        *p = VH
    case "hv":
        *p = HV
    case "vv":
        *p = VV
    default:
        err = errors.UnrecognizedMode(s, "polarization")
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
    }
    return
}
