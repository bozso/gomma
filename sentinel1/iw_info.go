package sentinel1

import (
    "fmt"
    "math"
    
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/geometry"
    
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/utils/params"
)

const (
    nMaxBurst = 10
)

type(
    IWInfo struct {
        nburst int
        extent common.LatLonRegion
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

    pPar := path.New("tmp").ToFile()
    pTOPSPar := pPar.AddExt("TOPS_par").ToFile()
    
    _, err = parCmd.Call(nil, file, nil, nil, pPar, nil, pTOPSPar)
    if err != nil {
        return
    }

    _info, err := burstCorners.Call(pPar, pTOPSPar)
    if err != nil {
        return
    }
    
    par, err := pPar.ToValid()
    if err != nil {
        return
    }
    defer par.Remove()
    
    TOPSPar, err := pTOPSPar.ToValid()
    if err != nil {
        return
    }
    defer TOPSPar.Remove()

    _TOPS, err := data.NewGammaParams(TOPSPar)
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
    
    rect, err := common.ParseRegion(info)
    if err != nil {
        return
    }

    return IWInfo{
        nburst: nburst,
        extent: rect,
        bursts: numbers,
    }, nil
}

func (s1 Zip) Info(dst path.Dir) (iws IWInfos, err error) {
    ext := s1.newExtractor(dst)
    if err = ext.Err(); err != nil {
        return
    }
    defer ext.Close()

    for ii := 1; ii < 4; ii++ {
        Annot := ext.Extract(annot, ii)
        
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

func inIWs(p geometry.LatLon, IWs IWInfos) bool {
    for _, iw := range IWs {
        if iw.extent.Contains(p) {
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
