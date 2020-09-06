package sentinel1

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/bozso/gotoolbox/path"

    //"github.com/bozso/gomma/data"
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/common"
    //"github.com/bozso/gomma/mli"
)

var dirPaths = [4]string{"slc", "rslc", "mli", "rmli"}

type (
    Zip struct {
        Path          path.ValidFile
        Safe          path.File
        pol           common.Pol
        Templates     templates
        date          date.Range
        
        mission       string
        dateStr       string
        mode          string
        productType   string
        resolution    string
        level         string
        productClass  string
        absoluteOrbit string
        DTID          string
        UID           string
    }
    
    Zips []*Zip
)

func NewZip(zipPath path.ValidFile) (s1 *Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"
    
    s1 = &Zip{
        Path: zipPath,
    }
    zipBase := zipPath.Base().String()
    
    s1.mission = strings.ToLower(zipBase[:3])
    s1.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    s1.date, err = date.Long.NewRange(start, stop)
    
    if err != nil {
        return
    }

    s1.mode = zipBase[4:6]
    safe := path.New(strings.ReplaceAll(zipBase, ".zip", ".SAFE")).ToFile()
    tpl := fmt.Sprintf(rexTemplate, s1.mission)

    s1.Templates = newTemplates(safe, tpl)

    s1.Safe = safe

    err = s1.pol.Set(zipBase[14:16])
    if err != nil {
        return
    }

    s1.productType = zipBase[7:10]
    s1.resolution = string(zipBase[10])
    s1.level = string(zipBase[12])
    s1.productClass = string(zipBase[13])
    s1.absoluteOrbit = zipBase[49:55]
    s1.DTID = strings.ToLower(zipBase[56:62])
    s1.UID = zipBase[63:67]

    return
}

/*
var parS1SLC = common.Must("par_S1_SLC")

func (s1 Zip) ImportSLC(dst path.Dir) (err error) {
    var _annot, _calib, _tiff, _noise string
    
    ext := s1.newExtractor(dst)
    if err = ext.Err(); err != nil {
        return
    }
    defer ext.Close()

    pol := s1.pol
    tab := s1.tabName("slc", pol)

    file, err := tab.Create()
    if err != nil {
        return
    }
    defer file.Close()

    for ii := 1; ii < 4; ii++ {
        _annot = ext.Extract(annot, ii)
        _calib = ext.Extract(calib, ii)
        _tiff  = ext.Extract(tiff, ii)
        _noise = ext.Extract(noise, ii)

        if err = ext.Err(); err != nil {
            return
        }

        iw := s1.GetIW("slc", pol, ii)

        _, err = parS1SLC.Call(_tiff, _annot, _calib, _noise, iw.ParFile,
            iw.DatFile, iw.TOPSpar)
        if err != nil {
            return
        }

        line := fmt.Sprintf("%s %s %s\n", iw.DatFile, iw.ParFile, iw.TOPSpar)

        _, err = file.WriteString(line)

        if err != nil {
            return
        }
    }

    return
}
*/

func (s1 Zip) Quicklook(dst path.Dir) (s path.ValidFile, err error) {
    var ext = s1.newExtractor(dst)
    if err = ext.Err(); err != nil {
        return
    }
    defer ext.Close()

    s = ext.Extract(quicklook, 0)
    err = ext.Err()

    return
}

func (s1 Zip) Date() time.Time {
    return s1.date.Center()
}

type ByDate Zips

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d ByDate) Less(i, j int) bool {
    return d[i].Date().Before(d[j].Date())
}
