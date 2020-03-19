package common

type Polarization int

const (
    VV Polarization = iota
    VH
    HV
    HH
)

func (p Polarization) String() (s string) {
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
