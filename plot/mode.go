package plot

import (

)

type Mode int

const (
    Byte Mode = iota
    CC
    Decibel
    Deform
    Height
    Linear
    MagPhase
    MagPhasePwr
    Power
    SingleLook
    Unwrapped
    Undefined
    MaximumMode
)

var modes = [...]Mode{Byte, CC, Decibel, Deform, Height, Linear,
    MagPhase, MagPhasePwr, Power, SingleLook, Unwrapped, Undefined}
