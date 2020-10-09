package plot

import (
    "github.com/bozso/gotoolbox/command"
    
    "github.com/bozso/gomma/settings"
)

type Service interface {
    Raster(Plottable, RasArgs) error
    Display(Plottable, DisArgs) error
}

type Commands [MaximumMode]command.Command

type RasterCommands Commands
type DisplayCommands Commands

func (r *RasterCommands) CommandSet(cmd settings.Commands) (err error) {
    for _, mode := range modes {
        r[mode], err = cmd.MustGet(RasterMode(mode).CommandName())
        if err != nil {
            break
        }
    }
    return
}

func (d *DisplayCommands) CommandSet(cmd settings.Commands) (err error) {
    for _, mode := range modes {
        d[mode], err = cmd.MustGet(DisplayMode(mode).CommandName())
        if err != nil {
            break
        }
    }
    return
}

type ServiceImpl struct {
    plotters [MaximumMode]Plotter
}

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

type RasterMode Mode

func (r RasterMode) CommandName() (s string) {
    switch Mode(r) {
    case Byte:
        s = "rasbyte"
    case CC:
        s = "rascc"
    case Decibel:
        s = "ras_dB"
    case Height:
        s = "rashgt"
    case Linear:
        s = "ras_linear"
    case MagPhase:
        s = "rasmph"
    case MagPhasePwr:
        s = "rasmph_pwr"
    case Power:
        s = "raspwr"
    case SingleLook:
        s = "rasSLC"
    /// @TODO: check out wether the following mappings are correct
    case Deform:
        s = "rasdt_pwr"
    case Unwrapped:
        s = "rasdt_pwr"
    }
    return
}

type DisplayMode Mode

func (d DisplayMode) CommandName() (s string) {
    switch Mode(d) {
    case Byte:
        s = "disbyte"
    case CC:
        s = "discc"
    case Decibel:
        s = "dis_dB"
    case Height:
        s = "dishgt"
    case Linear:
        s = "dis_linear"
    case MagPhase:
        s = "dismph"
    case MagPhasePwr:
        s = "dismph_pwr"
    case Power:
        s = "dispwr"
    case SingleLook:
        s = "disSLC"
    /// @TODO: check out wether the following mappings are correct
    case Deform:
        s = "disdt_pwr"
    case Unwrapped:
        s = "disdt_pwr"
    }
}
