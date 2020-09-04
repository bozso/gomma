package sentinel1

import (
    "io"
    "fmt"
    "bytes"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/common"
)

type SwathFlag int

const (
    AsListed SwathFlag = 0
)

type Noise int

const (
    ApplyCorrection Noise = iota
    NoCorrection
)

type ImportOptions struct {
    OPODDirectory path.Dir
    burstTable path.ValidFile
    SwathFlag
    Noise
    pol common.Pol
}

func (iop ImportOptions) ToArgs() (s string) {
    const template = "%s 0 0 \".\" 1 1"
    // TODO: actually implement
    // 0 = FCOMPLEX, 0 = swath_flag - as listed in burst_number_table_ref
    // "." = OPOD dir, 1 = intermediate files are deleted
    // 1 = apply noise correction
    return fmt.Sprintf(template, iop.pol)
}

type Importer struct {
    buf bytes.Buffer
    opArgs string
    burstTable path.ValidFile
    ZiplistFile path.File
}

func (iop ImportOptions) New(burstTable path.ValidFile) (im Importer) {
    im.burstTable, im.opArgs = burstTable, iop.ToArgs()
    return
}

var s1Import = common.Must("S1_import_SLC_from_zipfiles")

func (im Importer) Import(one, two *Zip) (err error) {
    err = im.WriteZiplist(one, two)
    if err != nil {
        return
    } 
    
    _, err = s1Import.Call(im.ZiplistFile, im.burstTable,
        im.opArgs)
    return
}

func (im Importer) WriteZiplist(one, two *Zip) (err error) {
    zipList, err := im.ZiplistFile.Create()
    if err != nil {
        return
    }
    defer zipList.Close()
    
    err = im.WriteZiplistTo(zipList, one, two)
    return 
}

func (im Importer) WriteZiplistTo(w io.Writer, one, two *Zip) (err error) {
    if two == nil {
        _, err = fmt.Fprintf(w, "%s\n", one.Path.String())
    } else {
        after := two.Date().After(one.Date())
        
        first, second := one, two
        
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
