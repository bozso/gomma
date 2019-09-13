package gamma


import (
    "math"
)

type(
    IWInfo struct {
		nburst int
		extent Rect
		bursts [nMaxBurst]float64
	}
	
    IWInfos [nIW]IWInfo

	S1IW struct {
		dataFile
		TOPS_par Params
	}
    
    IWs [nIW]S1IW

	S1SLC struct {
		date
        IWs IWs
		tab string
	}
)

const (
    nMaxBurst = 10
    nIW = 3
)


func NewIW(dat, par, TOPS_par string) (self S1IW, err error) {
    handle := Handler("NewS1SLC")
    
    self.dataFile, err = NewDataFile(dat, par)
    
    if err != nil {
        err = handle(err,
            "Failed to create DataFile with dat: '%s' and par '%s'!",
            dat, par)
        return
    }
    
    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }
    
    self.TOPS_par, err = NewGammaParam(TOPS_par)
    
    if err != nil {
        err = handle(err, "Failed to parse TOPS_parfile: '%s'!", TOPS_par)
        return
    }
    
    return self, nil
}

func makePoint(info Params, max bool) (ret Point, err error) {
	handle := Handler("makePoint")

	var tpl_lon, tpl_lat string

	if max {
		tpl_lon, tpl_lat = "Max_Lon", "Max_Lat"
	} else {
		tpl_lon, tpl_lat = "Min_Lon", "Min_Lat"
	}

	if ret.X, err = info.Float(tpl_lon); err != nil {
		err = handle(err, "Could not get Longitude value!")
        return
	}

	if ret.Y, err = info.Float(tpl_lat); err != nil {
		err = handle(err, "Could not get Latitude value!")
        return
	}

	return ret, nil
}

func iwInfo(path string) (ret IWInfo, err error) {
	handle := Handler("iwInfo")
    
    // num, err := conv.Atoi(str.Split(path, "iw")[1][0]);
    
    if len(path) == 0 {
        err = handle(nil, "path to annotation file is an empty string: '%s'",
            path)
        return
    }
    
	// Check(err, "Failed to retreive IW number from %s", path);

	par, err := TmpFile()

	if err != nil {
		err = handle(err, "Failed to create tmp file!")
        return
	}

	TOPS_par, err := TmpFile()

	if err != nil {
		err = handle(err, "Failed to create tmp file!")
        return
	}

	_, err = Gamma["par_S1_SLC"](nil, path, nil, nil, par, nil, TOPS_par)
    
	if err != nil {
		err = handle(err, "Could not import parameter files from '%s'!",
            path)
        return
	}
    
    _info, err := burstCorners(par, TOPS_par)
    
	if err != nil {
		err = handle(err, "Failed to parse parameter files!")
        return
	}
    
	info := FromString(_info, ":")
	TOPS, err := FromFile(TOPS_par, ":")

	if err != nil {
		err = handle(err, "Could not parse TOPS_par file!")
        return
	}

	nburst, err := TOPS.Int("number_of_bursts")

	if err != nil {
		err = handle(err, "Could not retreive number of bursts!")
        return
	}

	var numbers [nMaxBurst]float64
    
	for ii := 1; ii < nburst + 1; ii++ {
		tpl := fmt.Sprintf(burstTpl, ii)
        
		numbers[ii - 1], err = TOPS.Float(tpl)

		if err != nil {
			err = handle(err, "Could not get burst number: '%s'", tpl)
            return
		}
	}

	max, err := makePoint(info, true)

	if err != nil {
		err = handle(err, "Could not create max latlon point!")
        return
	}

	min, err := makePoint(info, false)

	if err != nil {
		err = handle(err, "Could not create min latlon point!")
        return
	}

	return IWInfo{nburst: nburst, extent: Rect{Min: min, Max: max},
		bursts: numbers}, nil
}

func (self *Point) inIWs(IWs IWInfos) bool {
	for _, iw := range IWs {
		if self.InRect(&iw.extent) {
			//log.Printf("%v in %v", *self, iw.extent)
            return true
		}
        //log.Printf("%v not in %v", *self, iw.extent)
	}
	return false
}

func (self IWInfos) contains(aoi AOI) bool {
	sum := 0
    
	for _, point := range aoi {
		if point.inIWs(self) {
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

func IWAbsDiff(one, two IWInfos) (float64, error) {
    sum := 0.0
    
    for ii := 0; ii < 3; ii++ {
        nburst1, nburst2 := one[ii].nburst, two[ii].nburst
        if nburst1 != nburst2 {
            return 0.0, fmt.Errorf(
            "In: IWInfos.AbsDiff: number of burst in first SLC IW%d (%d) " +
            "is not equal to the number of burst in the second SLC IW%d (%d)",
            ii + 1, nburst1, ii + 1, nburst2)
        }
        
        for jj := 0; jj < nburst1; jj++ {
            dburst := one[ii].bursts[jj] - two[ii].bursts[jj]
            sum += dburst * dburst
        }
    }
    
    return math.Sqrt(sum), nil
}
