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
    
    ifgOpt struct {
        Looks RangeAzimuth
        interact bool
        hgt string
        algo OffsetAlgo
    }
    
    AdaptFiltOpt struct {
        Offset               RangeAzimuth
        alpha, step, Frac    float64
        FFTWindow, CohWindow int
    }
)

const (
    Constant CohWeight = iota
    Gaussian

    IntensityCoherence OffsetAlgo = iota
    FingeVisibility
)

var (
    createOffset  = Gamma.must("create_offset")
    phaseSimOrb   = Gamma.must("phase_sim_orb")
    slcDiffIntf   = Gamma.must("SLC_diff_intf")
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
    
    self.qual, self.diffPar, self.qual, self.simUnw = 
    parfile, diffPar, quality, simUnw
    
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

func (self *IFG) AdaptFilt() (ret IFG, err error) {
    def adf(self, filt=None,, loff=0, nlines=0, wfrac=0.7):

    if step is None:
        step = fftwin / 8
    
    if filt is None:
        filt = self.datfile + ".filt"

    if cc is None:
        cc = self.datfile + ".cc"

    self.filt, self.cc = filt, cc
    
    rng = self.rng()
    
    gp.adf(self.dat, self.filt, self.cc, rng, alpha, fftwin, ccwin,
           step, loff, nlines, wfrac)


def coherence(self, slope_win=5, weight_type="gaussian", corr_thresh=0.4,
              box_lims=(3.0,9.0)):
    wgt_flag = IFG.cc_weights[weight_type]
    
    #log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    #log.info('Weight type is "%s"'.format(weight_type))
    
    width = ifg.rng()
    
    log.info("Estimating phase slope.", end=" ")
    gp.phase_slope(ifg.dat, slope, width, slope_win, corr_thresh)

    log.info("Calculating coherence.", end=" ")
    gp.cc_ad(ifg.dat, mli1, mli2, slope, None, ifg.cc, width, box_lims[0],
             box_lims[1], wgt_flag)

    
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

    
    def rascc(self):
        pass



