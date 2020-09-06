package service

import (
    "fmt"
    "bufio"
    "net/http"
    
    //"github.com/bozso/gotoolbox/path"

    //"github.com/bozso/emath/geometry"
    
    "github.com/bozso/gomma/date"
    "github.com/bozso/gomma/common"
    s1 "github.com/bozso/gomma/sentinel1"
)

type SentinelImport struct {
    Output
    Input
    MasterDate date.ShortTime  `json:"master_date"`
    Pol        common.Pol      `json:"polarization"`
}

var s1Import = common.Must("S1_import_SLC_from_zipfiles")

func (s *Sentinel1) DataImport(_ http.Request, ss *SentinelImport, _ *Empty) (err error) {
    const (
        tpl = "iw%d_number_of_bursts: %d\niw%d_first_burst: %f\niw%d_last_burst: %f\n"
        burst_table = "burst_number_table"
        ziplist = "ziplist"
    )
    
    defer ss.In.Close()
    zips, err := loadS1(ss.In)
    if err != nil {
        return
    }
    
    if !ss.MasterDate.IsSet() {
        return fmt.Errorf("master date is not set")
    }
    
    masterDate := date.Short.Format(ss.MasterDate.Time)
    var master *s1.Zip
    for _, s1zip := range zips {
        if date.Short.Format(s1zip.Date()) == masterDate {
            master = s1zip
        }
    }
    
    if master == nil {
        return fmt.Errorf("could not find master file, Sentinel 1 zipfile with date '%s' not found", masterDate)
    }
    
    var masterIW IWInfos
    if masterIW, err = master.Info(imp.CachePath); err != nil {
        return Handle(err, "failed to parse S1Zip data from master '%s'",
            master.Path)
    }
    
    fburst := NewWriterFile(burst_table);
    if err = fburst.Wrap(); err != nil {
        return
    }
    defer fburst.Close()
    
    fburst.WriteFmt("zipfile: %s\n", master.Path)
    
    nIWs := 0
    
    for ii, iw := range imp.IWs {
        min, max := iw.Min, iw.Max
        
        if min == 0 && max == 0 {
            continue
        }
        
        nburst := max - min
        
        if nburst < 0 {
            return Handle(nil, "number of burst for IW%d is negative, did " +
                "you mix up first and last burst numbers?", ii + 1)
        }
        
        IW := masterIW[ii]
        first := IW.bursts[min - 1]
        last := IW.bursts[max - 1]
        
        line := fmt.Sprintf(tpl, ii + 1, nburst, ii + 1, first, ii + 1, last)
        
        fburst.WriteString(line)
        nIWs++
    }
    
    if err = fburst.Wrap(); err != nil {
        return
    }
    
    // defer os.Remove(ziplist)
    
    slcDir := filepath.Join(imp.OutputDir, "SLC")

    if err = os.MkdirAll(slcDir, os.ModePerm); err != nil {
        return DirCreateErr.Wrap(err, slcDir)
    }
    
    pol, writer := imp.Pol, bufio.NewWriter(&imp.OutFile)
    defer imp.OutFile.Close()
    
    for _, s1zip := range zips {
        // iw, err := s1zip.Info(extInfo)
        
        date := date2str(s1zip, DShort)
        
        other := Search(s1zip, zips)
        
        err = toZiplist(ziplist, s1zip, other)
        
        if err != nil {
            return Handle(err, "could not write zipfiles to zipfile list file '%s'",
                ziplist)
        }
        
        if err != nil {
            return Handle(err, "failed to import zipfile '%s'", s1zip.Path)
        }
        
        base := fmt.Sprintf("%s.%s", date, pol)
        
        slc := S1SLC{
            Tab: base + ".SLC_tab",
            nIW: nIWs,
        }
        
        for ii := 0; ii < nIWs; ii++ {
            dat := fmt.Sprintf("%s.slc.iw%d", base, ii + 1)
            
            slc.IWs[ii].Dat = dat
            slc.IWs[ii].Params = NewGammaParam(dat + ".par")
            slc.IWs[ii].TOPS_par = NewGammaParam(dat + ".TOPS_par")
        }
        
        if slc, err = slc.Move(slcDir); err != nil {
            return
        }
        
        if _, err = writer.WriteString(slc.Tab); err != nil {
            return
        }
    }
    
    // TODO: save master idx?
    //err = SaveJson(path, meta)
    //
    //if err != nil {
    //    return Handle(err, "failed to write metadata to '%s'", path)
    //}
    
    return nil
}
