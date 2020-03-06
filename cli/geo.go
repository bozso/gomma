package cli

import (
    "strings"
    
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/dem"
    "github.com/bozso/gamma/geo"
)

type geoCode struct {
    InFile, OutFile data.File
    Mode            string
    dem.Lookup
    geo.CodeOpt
}

func (g *geoCode) SetCli(c *Cli) {
    c.Var(&g.Lookup, "lookup", "Lookup table file.")
    
    c.Var(&g.InFile, "infile", "Input datafile to geocode.")
    c.Var(&g.OutFile, "outfile", "Geocoded output datafile.")
    c.StringVar(&g.Mode, "mode", "",
        "Geocoding direction; from or to radar cordinates.")
    
    g.CodeOpt.SetCli(c)
}

func (c geoCode) Run() (err error) {
    mode := strings.ToUpper(c.Mode)
    
    switch mode {
    case "TORADAR", "RADAR":
        err = c.Lookup.geo2radar(c.InFile, c.OutFile, c.CodeOpt)
    case "TOGEO", "GEO":
        err = c.Lookup.radar2geo(c.InFile, c.OutFile, c.CodeOpt)
    default:
        err = UnrecognizedMode{name: "geocoding", got: mode}
    }
    
    return
}
