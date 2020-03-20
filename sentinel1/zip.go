package sentinel1

import (
    "fmt"
    "math"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/utils/params"
    "github.com/bozso/gomma/base"
)

var dirPaths = [4]string{"slc", "rslc", "mli", "rmli"}

type (
    Zip struct {
        Path path.ValidFile
        
        Path          string
        Root          string
        zipBase       string
        mission       string
        dateStr       string
        mode          string
        productType   string
        resolution    string
        Safe          string
        level         string
        productClass  string
        pol           string
        absoluteOrbit string
        DTID          string
        UID           string
        Templates     templates
        date          date.Range
    }
    
    Zips []*Zip
)

func NewZip(zipPath path.ValidFile, pol common.Pol) (s1 *Zip, err error) {
    const rexTemplate = "%s-iw%%d-slc-%%s-.*"

    
    s1.Path, s1.zipBase, s1.pol = zipPath, zipPath.Base(), pol
    zipBase := s1.zipBase.GetPath()

    s1.mission = strings.ToLower(zipBase[:3])
    s1.dateStr = zipBase[17:48]

    start, stop := zipBase[17:32], zipBase[33:48]

    s1.date, err = date.Long.NewRange(start, stop)
    
    if err != nil {
        return
    }

    s1.mode = zipBase[4:6]
    safe := path.New(strings.ReplaceAll(zipBase, ".zip", ".SAFE"))
    tpl := fmt.Sprintf(rexTemplate, s1.mission)

    s1.Templates = newTemplates(safe, tpl)

    s1.Safe = safe
    s1.productType = zipBase[7:10]
    s1.resolution = string(zipBase[10])
    s1.level = string(zipBase[12])
    s1.productClass = string(zipBase[13])
    s1.pol = zipBase[14:16]
    s1.absoluteOrbit = zipBase[49:55]
    s1.DTID = strings.ToLower(zipBase[56:62])
    s1.UID = zipBase[63:67]

    return s1, nil
}

func (s1 Zip) Names(mode, pol common.Pol) (p data.PathWithPar) {
    path := s1.Root.Join(mode)
    
    p.DatFile = path.Join(fmt.Sprintf("%s.%s", pol, mode)).ToFile()
    p.ParFile = dat.AddExt("par").ToFile()
    
    return 
}

func (s1 Zip) GetIW(mode string, pol common.Pol, ii int) (p IWPath) {
    slcPath := s1.Root.Join(mode)
    
    p = NewIW(slcPath.Join(fmt.Sprintf("iw%d_%s.%s", ii, pol, mode)))

    return
}

func (s1 Zip) SLC(pol string) (s SLC, err error) {
    const mode = "slc"
    tab := s1.tabName(mode, pol)

    exist, err := tab.Exist()
    if err != nil {
        return
    }

    if !exist {
        err = errors.WrapFmt(err, "tabfile '%s' does not exist", tab)
        return
    }

    for ii := 1; ii < 4; ii++ {
        iwp := s1.GetIW(mode, pol, ii)
        s.IWs[ii-1], err = iwp.Load()
    }

    s.Tab, s.nIW = tab, 3

    return
}

func (s1 Zip) MLI(mode, pol string, out *base.MLI, opt *base.MLIOpt) (err error) {
    slc, err := s1.SLC(pol)
    if err != nil {
        return
    }
    
    err = slc.MLI(out, opt)
    
    return
}

func (s1 Zip) tabName(mode, pol common.Pol) path.Path {
    return s1.Root.Join(mode, fmt.Sprintf("%s.tab", pol))
}

var parS1SLC = common.Must("par_S1_SLC")

func (s1 Zip) ImportSLC(dst string) (err error) {
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

func (s1 Zip) Quicklook(dst string) (s string, err error) {
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
    return date.Before(d[i], d[j])
}
