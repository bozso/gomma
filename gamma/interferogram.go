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
    CpxToReal  int
    
    IFG struct {
        dataFile
        DiffPar Params          `json:"diffparfile"`
        Quality   string        `json:"quality"`
        SimUnwrap string        `json:"simulated_unwrapped"`
        SLC1      SLC           `json:"slc1"`
        SLC2      SLC           `json:"slc2"`
        DeltaT    time.Duration `json:"-"`
    }
    
    Coherence struct {
        dataFile
    }
        
    IfgOpt struct {
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
)

const (
    IntensityCoherence OffsetAlgo = iota
    FingeVisibility
)

const (
    Real CpxToReal = iota
    Imaginary
    Intensity
    Magnitude
    Phase
)

var (
    createOffset  = Gamma.Must("create_offset")
    phaseSimOrb   = Gamma.Must("phase_sim_orb")
    slcDiffIntf   = Gamma.Must("SLC_diff_intf")
    adf           = Gamma.Must("adf")
    phaseSlope    = Gamma.Must("phase_slope")
    CCAdaptive    = Gamma.Must("cc_ad")

    CoherenceWeight = map[string]int {
        "constant": 0,
        "gaussian": 1,
    }
    
    IFGType = "IFG"
)

// TODO: check datatype of coherence file
func NewCoherence(dat, par string) (ret Coherence, err error) {
    ret.dataFile, err = NewDataFile(dat, par, Float)
    return
}

func NewIFG(dat, par, simUnw, diffPar, quality string) (ret IFG, err error) {
    if len(dat) == 0 {
        err = Handle(nil, "'dat' should not be an empty string")
        return
    }
    
    ret.Dat = dat
    
    base := NoExt(dat)
    
    if len(par) == 0 {
        par = base + ".off"
    }
    
    ret.Params = NewGammaParam(par)
    
    if len(simUnw) == 0 {
        simUnw = base + ".sim_unw"
    }
    
    ret.Quality, ret.SimUnwrap = quality, simUnw
    
    ret.DiffPar = NewGammaParam(diffPar)
    
    if ret.Rng, err = ret.rng(); err != nil {
        err = Handle(err, "failed to retreive range samples from '%s'", par)
        return
    }
    
    if ret.Azi, err = ret.azi(); err != nil {
        err = Handle(err, "failed to retreive azimuth lines from '%s'", par)
        return
    }
    
    ret.Dtype = FloatCpx
    
    return ret, nil
}
    
func FromSLC(slc1, slc2, ref *SLC, opt IfgOpt) (ret IFG, err error) {
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
    
    if ret, err = NewIFG(diff, off, "", "", ""); err != nil {
        err = Handle(err, "failed to create new interferogram struct")
        return
    }
    
    // TODO: Check date difference order
    ret.SLC1, ret.SLC2, ret.DeltaT = *slc1, *slc2, slc1.Time.Sub(slc2.Time)
    
    return ret, nil
}

var cpxToReal = Gamma.Must("cpx_to_real")

func (ifg IFG) ToReal(out string, mode CpxToReal) (ret dataFile, err error) {
    if ret, err = TmpDataFile(); err != nil {
        err = Handle(err, "failed to create temporary datafile")
        return
    }
    
    ret.RngAzi, ret.Dtype = ifg.RngAzi, Float
    
    Mode := 0
    
    switch (mode) {
    case Real:
        Mode = 0
    case Imaginary:
        Mode = 1
    case Intensity:
        Mode = 2
    case Magnitude:
        Mode = 3
    case Phase:
        Mode = 4
    default:
        return ret, Handle(nil, "Unrecognized mode!")
    }
    
    if _, err = cpxToReal(ifg.Dat, ret.Dat, ret.Rng, Mode); err != nil {
        err = Handle(err, "cpx_to_real failed")
        return
    }
    
    return ret, nil
}

var rasmph_pwr24 = Gamma.Must("rasmph_pwr24")

func (ifg IFG) Raster(opt RasArgs) error {
    err := opt.Parse(ifg)
    
    if err != nil {
        return Handle(err, "failed to parse IFG raster arguments")
    }
    
    cc := opt.Coh
    
    if len(cc) == 0 {
        _, err := rasmph_pwr24(opt.Datfile, opt.Sec, opt.Rng, opt.StartCpx,
                               opt.StartPwr, opt.Nlines, opt.Avg.Rng,
                               opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                               opt.Raster)
        if err != nil {
            return Handle(err, "failed to create raster for interferogram '%s'",
                ifg.Dat)
        }
    } else {
        
        _, err := rasmph_pwr24(opt.Datfile, opt.Sec, opt.Rng, opt.StartCpx,
                               opt.StartPwr, opt.Nlines, opt.Avg.Rng,
                               opt.Avg.Azi, opt.Scale, opt.Exp, opt.LR,
                               opt.Raster, cc, opt.StartCC, opt.Min)
        if err != nil {
            return Handle(err, "failed to create raster for interferogram '%s'",
                ifg.Dat)
        }
    }
    
    return nil
}

func (ifg IFG) rng() (int, error) {
    return ifg.Int("interferogram_width", 0)
}

func (ifg IFG) azi() (int, error) {
    return ifg.Int("interferogram_azimuth_lines", 0)
}

/*
 * TODO: remove?
func (ifg IFG) imgfmt() (string, error) {
    return "FCOMPLEX", nil
}
*/


func (self IFG) CheckQuality() (ret bool, err error) {
    qual := self.Quality
    
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

func (self IFG) AdaptFilt(opt AdaptFiltOpt) (ret IFG, cc Coherence, err error) {
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
    
    rng := self.Rng
    
    _, err = adf(self.Dat, ret.Dat, cc.Dat, rng, opt.alpha, opt.FFTWindow,
                 opt.cohWindow, step, opt.offset.Azi, opt.offset.Rng,
                 opt.frac)
    
    if err != nil {
        err = Handle(err, "adaptive filtering failed")
        return
    }
    
    return ret, cc, nil
}

func (self IFG) Coherence(opt CoherenceOpt) (ret Coherence, err error) {
    weightFlag := CoherenceWeight[opt.WeightType]
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width := self.Rng
    
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

/*
var rascc = Gamma.Must("rascc")

func (c Coherence) Raster(mli *MLI, opt IfgPlotOpt) error {
    err := opt.RasArgs.Parse(c)
    
    if err != nil {
        return Handle(err, "failed to parse plot arguments")
    }
    
    _, err = rascc(opt.Datfile, mli.Dat, opt.Rng, opt.StartCC, opt.StartPwr,
                   opt.Nlines, opt.Avg.Rng, opt.Avg.Azi,
                   opt.Min, opt.Max, opt.Scale,
                   opt.Exp, opt.LR, opt.Raster)
    
    return err
}
*/

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
