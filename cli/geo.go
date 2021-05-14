package cli

import (
//"strings"

//"github.com/bozso/gotoolbox/cli"

//"github.com/bozso/gomma/data"
//"github.com/bozso/gomma/dem"
//"github.com/bozso/gomma/geo"
)

/*
type geoCode struct {
    InFile, OutFile data.File
    Mode            string
    dem.Lookup
    geo.CodeOpt
}

func (g *geoCode) SetCli(c *cli.Cli) {
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
*/
