package gamma

import (
    "time"
)

type(
	dataFile struct {
		dat string
		Params
		date
        files []string
	}

    DataFile interface {
		Datfile() string
        Parfile() string
		Rng() int
		Azi() int
		Int() int
		Float() float64
	}
)

func NewGammaParam(path string) (Params, error) {
    return FromFile(path, ":")
}

func NewDataFile(dat, par string) (ret dataFile, err error) {
    handle := Handler("NewDatfile")
    ret.dat = dat
    
    if len(dat) == 0 {
        err = handle(err, "'dat' should not be an empty string: '%s'", dat)
    }
    
    if len(par) == 0 {
        par = dat + ".par"
    }
    
    ret.Params, err = NewGammaParam(par)
    
    if err != nil {
        err = handle(err, "Failed to parse gamma parameter file: '%s'", par)
        return
    }
    
    ret.files = []string{dat, par}
    
    return ret, nil
}

func (self *dataFile) Exist() (ret bool, err error) {
    for _, file := range self.files {
        exist, err = Exist(file)
        
        if err != nil {
            err = fmt.Errorf("Stat on file '%s' failed!\nError: %w!",
                file, err)
            return
        }
        
        if !exist {
            return false, nil
        }
    }
    return true, nil
}

func (self *dataFile) Datfile() string {
    return self.dat
}

func (self *dataFile) Parfile() string {
    return self.par
}

func (self *dataFile) Rng() (int, error) {
	return self.Int("range_samples")
}

func (self *dataFile) Azi() (int, error) {
	return self.Int("azimuth_samples")
}

func (self *dataFile) imgFormat() (string, error) {
	return self.Par("image_format")
}

func (self *dataFile) Date() time.Time {
	return self.center
}
