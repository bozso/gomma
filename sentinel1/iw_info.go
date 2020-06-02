package sentinel1

import (
    "github.com/bozso/gotoolbox/path"
    
    "github.com/bozso/gomma/common"
)

const (
    nMaxBurst = 10
)

type(
    IWInfo struct {
        nburst int
        extent common.Rectangle
        bursts [nMaxBurst]float64
    }
    
    IWInfos [maxIW]IWInfo
)

var (
    parCmd = common.Must("par_S1_SLC")
    burstCorners = common.Select("ScanSAR_burst_corners", "SLC_burst_corners")
)

func iwInfo(file path.ValidFile) (iw IWInfo, err error) {
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    
    // Check(err, "Failed to retreive IW number from %s", path);

    par := "tmp"
    TOPS_par := par + ".TOPS_par"
    
    defer os.Remove(par)
    defer os.Remove(TOPS_par)

    _, err = parCmd.Call(nil, file.GetPath(), nil, nil, par, nil, TOPS_par)
    if err != nil {
        return
    }

    _info, err := burstCorners.Call(par, TOPS_par)
    if err != nil {
        return
    }

    _TOPS, err := data.NewGammaParams(TOPS_par)
    if err != nil {
        return
    }
    TOPS := _TOPS.ToParser()
    
    nburst, err := TOPS.Int("number_of_bursts", 0)

    if err != nil {
        return
    }

    var numbers [nMaxBurst]float64

    const burstTpl = "burst_asc_node_%d"

    for ii := 1; ii < nburst+1; ii++ {
        tpl := fmt.Sprintf(burstTpl, ii)

        numbers[ii-1], err = TOPS.Float(tpl, 0)

        if err != nil {
            return
        }
    }
    
    info := params.FromString(_info, ":").ToParser()
    
    rect, err := common.ParseRectangle(info)
    if err != nil {
        return
    }

    return IWInfo{
        nburst: nburst,
        extent: rect,
        bursts: numbers,
    }, nil
}

func (s1 Zip) Info(dst string) (iws IWInfos, err error) {
    ext := s1.newExtractor(dst)
    if err = ext.Err(); err != nil {
        return
    }
    defer ext.Close()

    var Annot string
    for ii := 1; ii < 4; ii++ {
        Annot = ext.Extract(annot, ii)
        
        if err = ext.Err(); err != nil {
            return
        }
        
        iws[ii-1], err = iwInfo(Annot)
        
        if err != nil {
            err = common.ParseFail(Annot, err).ToRetreive("IW information")
            return
        }
    }
    
    return
}

func inIWs(p common.Point, IWs IWInfos) bool {
    for _, iw := range IWs {
        if p.InRect(iw.extent) {
            return true
        }
    }
    return false
}

func (iw IWInfos) contains(aoi common.AOI) bool {
    sum := 0

    for _, point := range aoi {
        if inIWs(point, iw) {
            sum++
        }
    }
    return sum == 4
}

func diffBurstNum(burst1, burst2 float64) int {
    dburst := burst1 - burst2
    diffSqrt := math.Sqrt(dburst)

    return int(dburst + 1.0 + (dburst / (0.001 + diffSqrt)) * 0.5)
}

func checkBurstNum(one, two IWInfos) bool {
    for ii := 0; ii < 3; ii++ {
        if one[ii].nburst != two[ii].nburst {
            return true
        }
    }
    return false
}

func IWAbsDiff(one, two IWInfos) (sum float64, err error) {
    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            err = fmt.Errorf(
                "number of burst in first SLC IW%d (%d) is not equal to " + 
                "the number of burst in the second SLC IW%d (%d)",
                ii + 1, nburst1, ii + 1, nburst2)
            return
        }

        for jj := 0; jj < nburst1; jj++ {
            dburst := one[ii].bursts[jj] - two[ii].bursts[jj]
            sum += dburst * dburst
        }
    }

    return math.Sqrt(sum), nil
}
