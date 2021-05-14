package plot

import (
	"strings"

	"github.com/bozso/gotoolbox/command"
	"github.com/bozso/gotoolbox/enum"

	"github.com/bozso/gomma/settings"
)

var plotType = enum.NewStringSet("Raster", "Display").EnumType("plot.Type")

type Type int

const (
	Raster Type = iota
	Display
)

func (t Type) String() (s string) {
	switch t {
	case Raster:
		s = "raster"
	case Display:
		s = "display"
	default:
	}
	return
}

func (t *Type) Set(s string) (err error) {
	switch strings.ToLower(s) {
	case "raster":
		*t = Raster
	case "display":
		*t = Display
	default:
		err = plotType.UnknownElement(s)
	}
	return
}

func (t *Type) UnmarshalJSON(b []byte) (err error) {
	err = t.Set(string(b))
	return
}

type Plotter interface {
	Raster(RasterOptions) error
	//Display(DisplayOptions) error
}

type CommandNames struct {
	Raster, Display string
}

type CommandNamer interface {
	CommandNames() CommandNames
}

func (m Mode) CommandNames() (c CommandNames) {
	switch m {
	case Byte:
		c.Raster, c.Display = "rasbyte", "disbyte"
	case CC:
		c.Raster, c.Display = "rascc", "discc"
	case Decibel:
		c.Raster, c.Display = "ras_dB", "dis_dB"
	case Height:
		c.Raster, c.Display = "rashgt", "dishgt"
	case MagPhase:
		c.Raster, c.Display = "rasmph", "dismph"
	case MagPhasePwr:
		c.Raster, c.Display = "rasmph_pwr", "dismph_pwr"
	case Power:
		c.Raster, c.Display = "raspwr", "dispwr"
	case SingleLook:
		c.Raster, c.Display = "rasSLC", "disSLC"
	/// TODO: check out whether the following mappings are correct
	case Deform, Unwrapped:
		c.Raster, c.Display = "rasdt_pwr", "disdt_pwr"
	}
	return
}

type PlotCommand struct {
	raster, display command.Command
}

func (m Mode) NewPlotCommand(c settings.Commands) (p PlotCommand, err error) {
	d := m.CommandNames()

	p.raster, err = c.Get(d.Raster)
	if err != nil {
		return
	}

	p.display, err = c.Get(d.Display)
	return
}
