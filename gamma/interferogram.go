package gamma

import (
    "os"
    "time"
    "log"
    bio "bufio"
    str "strings"
    conv "strconv"
)

type (
    CohWeight int
    OffsetAlgo int
    
    IFG struct {
        dataFile
        diffPar Params
        quality, simUnwrap string
        deltaT time.Duration
        slc1, slc2 SLC
    }
    
    Coherence struct {
        dataFile
    }
    
    ifgPlotOpt struct {
        rasArgs
        cc *Coherence
        startCC, startPwr, startCpx int
        Range minmax 
    }
    
    ifgOpt struct {
        Looks RngAzi
        interact bool
        hgt string
        algo OffsetAlgo
    }
    
    AdaptFiltOpt struct {
        offset               RngAzi
        alpha, step, frac    float64
        FFTWindow, cohWindow int
    }
    
    CoherenceOpt coherence
)

const (
    IntensityCoherence OffsetAlgo = iota
    FingeVisibility
)

var (
    createOffset  = Gamma.must("create_offset")
    phaseSimOrb   = Gamma.must("phase_sim_orb")
    slcDiffIntf   = Gamma.must("SLC_diff_intf")
    adf           = Gamma.must("adf")
    phaseSlope    = Gamma.must("phase_slope")
    CCAdaptive    = Gamma.must("cc_ad")

    CoherenceWeight = map[string]int {
        "constant": 0,
        "gaussian": 1,
    }
)

func NewCoherence(dat, par string) (ret Coherence, err error) {
    ret.dataFile, err = NewDataFile(dat, par, "par")
    return
}

func NewIFG(dat, par, simUnw, diffPar, quality string) (self IFG, err error) {
    if len(dat) == 0 {
        err = Handle(nil, "'dat' should not be an empty string")
        return
    }
    
    self.Dat = dat
    
    base := NoExt(dat)
    
    if len(par) == 0 {
        par = base + ".off"
    }
    
    self.Params = NewGammaParam(par)
    
    if len(simUnw) == 0 {
        simUnw = base + ".sim_unw"
    }
    
    self.quality, self.simUnwrap = quality, simUnw
    
    self.diffPar = NewGammaParam(diffPar)
    
    self.files = []string{}
    
    return self, nil
}
    

func (ifg *IFG) Move(dir string) error {
    err := Move(&ifg.dataFile.Dat, dir)
    
    if err != nil {
        return Handle(err, "failed to move IFG datafile")
    }
    
    err = Move(&ifg.dataFile.Par, dir)
    
    if err != nil {
        return Handle(err, "failed to move IFG parfile")
    }
    
    err = Move(&ifg.diffPar.Par, dir)
    
    if err != nil {
        return Handle(err, "failed to move IFG diff file")
    }
    
    err = Move(&ifg.quality, dir)
    
    if err != nil {
        return Handle(err, "failed to move IFG quality file")
    }
    
    err = Move(&ifg.simUnwrap, dir)
    
    if err != nil {
        return Handle(err, "failed to move IFG simulated unwrapped phase file")
    }
    
    return nil
}

func FromSLC(slc1, slc2, ref *SLC, opt ifgOpt) (ret IFG, err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    // TODO: figure out where IFGs will be placed
    off := "%s.off"
    simUnw := "%s.sim_unw"
    diff := "%s.diff"
    
    par1, par2 := slc1.Par, slc2.Par
    
    _, err = createOffset(par1, par2, off, opt.algo, rng, azi, inter)
    
    if err != nil {
        err = Handle(err, "failed to create offset table")
        return
    }
    
    slcRefPar := ""
    
    if ref != nil {
        slcRefPar = ref.Par
    }
    
    _, err = phaseSimOrb(par1, par2, off, opt.hgt, simUnw, slcRefPar,
                         nil, nil, 1)
    
    if err != nil {
        err = Handle(err, "failed to create simulated orbital phase")
        return
    }
    
    dat1, dat2 := slc1.Datfile(), slc2.Datfile()
    _, err = slcDiffIntf(dat1, dat2, par1, par2, off,
                         simUnw, diff, rng, azi, 0, 0)
    
    if err != nil {
        err = Handle(err, "failed to create differential interferogram")
        return
    }
    
    ret, err = NewIFG(diff, off, "", "", "")
    
    if err != nil {
        err = Handle(err, "failed to create new interferogram struct")
        return
    }
    
    // TODO: Check date difference order
    ret.slc1, ret.slc2, ret.deltaT = *slc1, *slc2, slc1.Center().Sub(slc2.Center())
    
    return ret, nil
}


func (opt *ifgPlotOpt) Parse(ifg *IFG) error {
    err := opt.rasArgs.Parse(ifg)
    
    if err != nil {
        return err
    }
    
    if opt.Range.Min == 0.0 {
        opt.Range.Min = 0.1
    }
    
    if opt.Range.Max == 0.0 {
        opt.Range.Min = 0.9
    }
    
    if opt.startCC == 0 {
        opt.startCC = 1
    }
    
    if opt.startPwr == 0 {
        opt.startPwr = 1
    }
    
    if opt.startCpx == 0 {
        opt.startCpx = 1
    }
    
    return nil
}

var rasmph_pwr24 = Gamma.must("rasmph_pwr24")

func (ifg *IFG) Raster(mli string, opt ifgPlotOpt) error {
    err := opt.Parse(ifg)
    
    if err != nil {
        return Handle(err, "failed to parse IFG raster arguments")
    }
    
    
    
    cc := opt.cc
    
    if cc == nil {
        _, err := rasmph_pwr24(opt.Datfile, mli, opt.Rng, opt.startCpx,
                               opt.startPwr, opt.Nlines, opt.Avg.Rng,
                               opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                               opt.raster)
        if err != nil {
            return Handle(err, "failed to create raster for interferogram '%s'",
                ifg.Dat)
        }
    } else {
        
        _, err := rasmph_pwr24(opt.Datfile, mli, opt.Rng, opt.startCpx,
                               opt.startPwr, opt.Nlines, opt.Avg.Rng,
                               opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                               opt.raster, *cc, opt.startCC, opt.Range.Min)
        if err != nil {
            return Handle(err, "failed to create raster for interferogram '%s'",
                ifg.Dat)
        }
    }
    
    return nil
}

func (ifg *IFG) Rng() (int, error) {
    return ifg.Int("interferogram_width", 0)
}

func (ifg *IFG) Azi() (int, error) {
    return ifg.Int("interferogram_azimuth_lines", 0)
}

func (ifg *IFG) ImageFormat() (string, error) {
    return "FCOMPLEX", nil
}

func (self *IFG) CheckQuality() (ret bool, err error) {
    qual := self.quality
    
    file, err := os.Open(qual)
    
    if err != nil {
        err = Handle(err, "failed to open file '%s'", qual)
        return
    }
    
    defer file.Close()
    
    scanner := bio.NewScanner(file)
    
    offs := 0.0
    
    var diff float64
    
    for scanner.Scan() {
        line := scanner.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := str.Fields(line)
        
        if split[0] == "azimuth_pixel_offset" {
            diff, err = conv.ParseFloat(split[1], 64)
            
            if err != nil {
                err = Handle(err, "failed to parse: '%s' into float64",
                    split[1])
                return
            }
            
            offs += diff
        }
    }
    
    log.Printf("Sum of azimuth offsets in %s is %f pixel.\n", qual, offs)
    
    if offs > 0.0 || offs < 0.0 {
        ret = true
    } else {
        ret = false
    }
    
    return ret, nil
}

func (self *IFG) AdaptFilt(opt AdaptFiltOpt) (ret IFG, cc Coherence, err error) {
    step := float64(opt.FFTWindow) / 8.0
    
    if opt.step > 0.0 {
        step = opt.step
    }
    
    // TODO: figure out the name of the output files
    ret, err = NewIFG(self.Dat + ".filt", "", "", "", "")
    
    if err != nil {
        err = Handle(err, "failed to create new interferogram struct")
        return
    }
    
    cc, err = NewCoherence("", "")
    
    if err != nil {
        err = Handle(err, "failed to create new dataFile struct")
        return
    }
    
    /*
    if Empty(filt):
        filt = 
    
    if empty(cc is None:
        cc = self.datfile + ".cc"
    */
    
    rng, err := self.Rng()
    
    if err != nil {
        err = Handle(err, "failed to retreive range samples")
        return
    }
    
    _, err = adf(self.Dat, ret.Dat, cc.Dat, rng, opt.alpha, opt.FFTWindow,
                 opt.cohWindow, step, opt.offset.Azi, opt.offset.Rng,
                 opt.frac)
    
    if err != nil {
        err = Handle(err, "adaptive filtering failed")
        return
    }
    
    return ret, cc, nil
}

func (self *IFG) Coherence(opt CoherenceOpt) (ret Coherence, err error) {
    weightFlag := CoherenceWeight[opt.WeightType]
    var width int
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width, err = self.Rng()
    
    if err != nil {
        err = Handle(err, "failed to retreive range samples")
        return
    }
    
    log.Printf("Estimating phase slope. ")
    
    // TODO: figure out name
    slope := ".cpx"
    
    // parameters: xmin, xmax, ymin, ymax not yet given
    _, err = phaseSlope(self.Dat, slope, opt.SlopeWindow,
                        opt.SlopeCorrelationThresh)
    
    if err != nil {
        err = Handle(err, "failed to calculate phase slope")
        return
    }

    log.Printf("Calculating coherence. ")
    
    mli1, mli2 := "", ""
    
    _, err = CCAdaptive(self.Dat, mli1, mli2, slope, nil, ret.Dat, width,
                        opt.Box.Min, opt.Box.Max, weightFlag)
    
    if err != nil {
        err = Handle(err, "adaptive filtering failed")
        return
    }
    
    return ret, nil
}

var rascc = Gamma.must("rascc")

func (c *Coherence) Raster(mli *MLI, opt ifgPlotOpt) error {
    err := opt.rasArgs.Parse(c)
    
    if err != nil {
        return Handle(err, "failed to parse plot arguments")
    }
    
    _, err = rascc(opt.Datfile, mli.Dat, opt.Rng, opt.startCC, opt.startPwr,
                   opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                   opt.Range.Min, opt.Range.Max, opt.Scale,
                   opt.Exp, opt.LR, opt.raster)
    
    return err
}

/*
def raster(self, start_cpx=1, start_pwr=1, start_cc=1, cc_min=0.2,
           **kwargs):
    mli = kwargs.pop("mli")
    
    args = DataFile.parse_ras_args(self, **kwargs)
    
    if self.cc is None:
        gp.rasmph_pwr24(args["datfile"], mli.dat, args["rng"],
                        start_cpx, start_pwr, args["nlines"],
                        args["arng"], args["aazi"], args["scale"],
                        args["exp"], args["LR"], args["raster"])
    else:
        gp.rasmph_pwr24(args["datfile"], mli.dat, args["rng"],
                        start_cpx, start_pwr, args["nlines"],
                        args["arng"], args["aazi"], args["scale"],
                        args["exp"], args["LR"], args["raster"],
                        self.cc, start_cc, cc_min)


*/
