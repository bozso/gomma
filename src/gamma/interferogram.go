package gamma

import (
    bio "bufio"
)

type (
    CohWeight int
    OffsetAlgo int
    
    IFG struct {
        dataFile
        diffPar Params
        qual, coherence, simUnwrap string
        deltaT time.Duration
        slc1, slc2 SLC
    }
    
    Coherence struct {
        dataFile
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

func NewIFG(dat, par, simUnw, diffPar, quality string) (self IFG, err error) {
    handle := Handler("NewIFG")
    
    if len(dat) == 0 {
        err = handle(nil, "'dat' should not be an empty string!")
    }
    
    self.dat = dat
    
    base := NoExt(dat)
    
    if length(par) == 0 {
        par = base + ".off"
    }
    
    self.Params, err = NewGammaParams(par)
    
    if err != nil {
        err = handle(err, "Could no parse parameter file: '%s'!", par)
        return
    }
    
    if len(simUnw) == 0 {
        simUnw = base + ".sim_unw"
    }
    
    self.qual, self.diffPar, self.simUnw = 
    quality, diffPar, simUnw
    
    return ret, nil
}
    
    
func FromSLC(slc1, slc2, ref *SLC, opt ifgOpt) (self IGF, err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Range, opt.Looks.Azimuth
    
    // TODO: figure out where IFGs will be placed
    off = "%s.off"
    sim_unw = "%s.sim_unw"
    diff = "%s.diff"
    
    _, err = createOffset(slc1.par, slc2.par, off, opt.algo, rng, azi, inter)
    
    slcRefPar := ""
    
    if ref != nil {
        slcRefPar = ref.par
    }
    _, err = phaseSimOrb(slc1.par, slc2.par, off, opt.hgt, simUnw, slcRefPar,
                         nil, nil, 1)
    
    _, err = slcDiffIntf(slc1.dat, slc2.dat, slc1.par, slc2.par, off,
                         simUnw, diff, rng, azi, 0, 0)
    
    ret = NewIFG(diff, off)
    ret.dat, ret.slc1, ret.slc2, ret.deltaT = diff, slc1, slc2
    
    return ret, nil
}

/*
TODO: translate these functions
def rng(self):
    return self.int("interferogram_width")

def azi(self):
    return self.int("interferogram_azimuth_lines")

def img_fmt(self):
    return "FCOMPLEX"
*/



func (self *IFG) CheckQuality() (ret bool, err error) {
    handle := Handler("IFG:CheckQuality")
    qual = self.qual
    
    file, err := os.Create(qual)
    
    if err != nil {
        err = handle(err, "Could not open file: '%s'!", qual)
        return
    }
    
    defer file.Close()
    
    scaner := bio.NewScanner(file)
    
    offs := 0.0
    
    for scaner.Scan() {
        line := scanner.Text()
        if str.HasPrefix(line, "azimuth_pixel_offset") {
            split := str.Split(line, " ")[1]
            
            diff, err := str.ParseFloat(split, 64)
            
            if err != nil {
                err = handle(err, "Could no parse: '%s' into float64!", split)
                return
            }
            
            offs += diff
        }
    }
    
    log.Printf("Sum of azimuth offsets in %s is %f pixel." qual, offs)
    
    if isclose(offs, 0.0) {
        ret = true
    } else {
        ret = false
    }
    
    return ret, nil
}

func (self *IFG) AdaptFilt() (ret IFG, cc Coherence, err error) {
    handle := Handler("IFG.AdaptFilt")
    step := opt.FFTWin / 8.0
    
    if opt.step > 0.0 {
        step = opt.step
    }
    
    filt = NewIFG(self.dat + ".filt")
    cc = NewDataFile()
    if filt is None:
        filt = 

    if cc is None:
        cc = self.datfile + ".cc"

    rng, err := self.Rng()
    
    if err != nil {
        err = handle(err, "Could not retreive range samples!")
        return
    }
    
    _, err := adf(self.dat, filt, cc, rng, opt.alpha, opt.FFTWindow,
                  opt.cohWindow, step, opt.offset.Azi, opt.offset.Rng,
                  opt.frac)
    
    if err != nil {
        err = handle(err, "Adaptive filtering failed!")
        return
    }
}

func (self *IFG) Coherence(opt *CoherenceOpt) (ret Coherence, err error) {
    weightFlag := CoherenceWeight[opt.WeightType]
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width = self.Rng()
		WeightType             string
		Box                    minmax
		SlopeCorrelationThresh float64
		SlopeWindow            int
    
    log.Printf("Estimating phase slope. ")
    
    // TODO: figure out name
    slope := ".cpx"
    
    // parameters: xmin, xmax, ymin, ymax not yet given
    _, err = phaseSlope(self.dat, slope, opt.SlopeWindow,
                        opt.SlopeCorrelationThresh)

    log.Printf("Calculating coherence. ")
    _, err = CCAdaptive(self.dat, mli1, mli2, slope, nil, ret.dat, width,
                        opt.Box.min, opt.Box.max, weightFlag)

/*
func (self *IFG) raster(args RasArgs) {
    args = parseRasArgs(args)
    
    
}
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