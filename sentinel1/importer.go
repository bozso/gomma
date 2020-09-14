package sentinel1

import (
    "io"
    "fmt"
    //"bytes"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/settings"
)

type SwathFlag int

const (
    AsListed SwathFlag = iota
    One
    Two
    Three
    OneTwo
    TwoThree
)

func (sf SwathFlag) String() (s string) {
    switch sf {
    case AsListed:
        s = "0"
    case One, Two, Three:
        s = fmt.Sprintf("%d", sf)
    case OneTwo:
        s = "4"
    case TwoThree:
        s = "5"
    default:
        s = "-"
    }
    return
}

type Noise int

const (
    ApplyCorrection Noise = iota
    NoCorrection
)

func (n Noise) String() (s string) {
    switch n {
    case ApplyCorrection:
        s = "1"
    case NoCorrection:
        s = "2"
    default:
        s = "-"
    }
    return
}

type ImportOptions struct {
    OPODDirectory string
    burstTable string
    dtype data.Type
    SwathFlag
    Noise
    pol common.Pol
}

var defaultImporter = ImportOptions{
    OPODDirectory: ".",
    burstTable: "",
    SwathFlag: AsListed,
    Noise: ApplyCorrection,
    pol: common.AllPolarisation,
}

func DefaultImporter() (io ImportOptions) {
    return defaultImporter
}

func (io ImportOptions) ToArgs() (s string, err error) {
    var buf strings.Buffer
    
    switch d := io.dtype; d {
    case data.FloatCpx:
        dt = "0"
    case data.ShortCpx:
        dt = "1"
    default:
        err = d.WrongType("Sentinel1 import")
        return
    }
    
    burstTable, err := path.New(io.burstTable).ToValidFile()
    if err != nil {
        return
    }
    
    opod, err := path.New(io.OPODDirectory).ToDir()
    if err != nil {
        return
    }
    
    s = fmt.Sprintf("%s %s %d %s", io.pol, dt, io.SwathFlag,
        opod, io.Cleaning, io.Noise)
    return
}

func (io ImportOptions) New(c settings.Commands) (im Importer, err error) {
    im.command, err = c.Get("S1_import_SLC_from_zipfiles")
    if err != nil {
        return
    }
    
    im.opArgs, err = io.ToArgs()
    return
}

type Importer struct {
    command settings.Command
    opArgs string
    burstTable path.ValidFile
}

func (im Importer) Import(one, two *Zip) (err error) {
    err = im.WriteZiplist(one, two)
    if err != nil {
        return
    } 
    
    _, err = im.command.Call(im.ZiplistFile, im.opArgs)
    return
}

func (im Importer) WriteZiplist(one, two *Zip) (err error) {
    zipList, err := im.ZiplistFile.Create()
    if err != nil {
        return
    }
    defer zipList.Close()
    
    err = im.FormatZiplist(zipList, one, two)
    return 
}

func (im Importer) FormatZiplist(w io.Writer, one, two *Zip) (err error) {
    if two == nil {
        _, err = fmt.Fprintf(w, "%s\n", one.Path.String())
    } else {
        first, second := one, two

        after := two.Date().After(one.Date())
        
        if !after {
            first, second = second, first
        }
        
        _, err = fmt.Fprintf(w, "%s\n", one.Path.String())
        if err != nil {
            return
        }
        
        _, err = fmt.Fprintf(w, "%s\n", second.Path.String())
    }
    
    return
}
