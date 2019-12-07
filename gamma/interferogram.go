package gamma

import (
    "os"
    "time"
    "log"
    "bufio"
    "strings"
    "strconv"
)

type (
    CohWeight int
    OffsetAlgo int
    CpxToReal  int
    
    Coherence struct {
        DatFile
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
)

type IFG struct {
    DatParFile
    DiffPar   Params        `json:"diffparfile"`
    Quality   string        `json:"quality"`
    SimUnwrap string        `json:"simulated_unwrapped"`
    DeltaT    time.Duration `json:"-"`
}

func NewIFG(dat, off, diffpar string) (ret IFG, err error) {
    if ret.DatParFile, err = NewDatParFile(dat, off, "off", FloatCpx);
       err != nil {
        return
    }
    
    if len(diffpar) == 0 {
        diffpar = dat + ".diff_par"
    }
    
    ret.DiffPar = NewGammaParam(diffpar)
    
    return ret, nil
}

func TmpIFG() (ret IFG, err error) {
    if ret.DatParFile, err = TmpDatParFile("diff", "off", FloatCpx); err != nil {
        return
    }
    
    ret.DiffPar = Params{Par: ret.Dat + ".par", Sep: ":"}
    ret.SimUnwrap = ret.Dat + ".sim_unw"
    
    return ret, nil
}

func (i IFG) jsonMap() JSONMap {
    ret := i.DatParFile.jsonMap()
    
    ret["quality"] = i.Quality
    ret["diffparfile"] = i.DiffPar
    ret["simulated_unwrapped"] = i.SimUnwrap
    
    return ret
}

func (i IFG) Move(dir string) (ret IFG, err error) {
    if ret.DatParFile, err = i.DatParFile.Move(dir); err != nil {
        return
    }
    
    if ret.DiffPar.Par, err = Move(i.DiffPar.Par, dir); err != nil {
        return
    }
    ret.DiffPar.Sep = ":"
    
    if len(i.SimUnwrap) > 0 {
        if ret.SimUnwrap, err = Move(i.SimUnwrap, dir); err != nil {
            return
        }
    }
    
    if len(i.Quality) > 0 {
        if ret.Quality, err = Move(i.Quality, dir); err != nil {
            return
        }
    }
    return ret, nil
}

func (i *IFG) FromJson(m JSONMap) (err error) {
    if err = i.DatParFile.FromJson(m); err != nil {
        return
    }
    
    if i.DType != FloatCpx {
        err = TypeMismatchError{ftype:"IFG", expected:"complex",
            DType:i.DType}
        return
    }
    
    if i.Quality, err = m.String("quality"); err != nil {
        err = Handle(err, "failed to retreive quality file")
        return
    }
    
    if i.DiffPar.Par, err = m.String("diffparfile"); err != nil {
        err = Handle(err, "failed to diffparfile")
        return
    }
    i.DiffPar.Sep = ":"
    
    
    if i.SimUnwrap, err = m.String("simulated_unwrapped"); err != nil {
        err = Handle(err, "failed to simulated unwrapped datafile")
        return
    }
    
    return nil
}
    
func FromSLC(slc1, slc2, ref *SLC, opt IfgOpt) (ret IFG, err error) {
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    par1, par2 := slc1.Par, slc2.Par
    
    _, err = createOffset(par1, par2, ret.Par, opt.algo, rng, azi, inter)
    
    if err != nil {
        err = Handle(err, "failed to create offset table")
        return
    }
    
    slcRefPar := "-"
    
    if ref != nil {
        slcRefPar = ref.Par
    }
    
    if ret, err = TmpIFG(); err != nil {
        return
    }
    
    _, err = phaseSimOrb(par1, par2, ret.Par, opt.hgt, ret.SimUnwrap, slcRefPar,
                         nil, nil, 1)
    
    dat1, dat2 := slc1.Dat, slc2.Dat
    _, err = slcDiffIntf(dat1, dat2, par1, par2, ret.Par,
                         ret.SimUnwrap, ret.DiffPar, rng, azi, 0, 0)
    
    
    if err = ret.Parse(); err != nil {
        return
    }
    
    // TODO: Check date difference order
    ret.DeltaT = slc1.Time.Sub(slc2.Time)
    
    return ret, nil
}

var cpxToReal = Gamma.Must("cpx_to_real")

func (ifg IFG) ToReal(mode CpxToReal) (ret DatFile, err error) {
    if ret, err = TmpDatFile("real", Float); err != nil {
        return
    }
    ret.URngAzi = ifg.URngAzi
    
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
        return
    }
    
    return ret, nil
}

var rasmph_pwr24 = Gamma.Must("rasmph_pwr24")

func (ifg IFG) Raster(opt RasArgs) error {
    opt.Mode = MagPhasePwr
    return ifg.Raster(opt)
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
    
    scanner := bufio.NewScanner(file)
    
    offs := 0.0
    
    var diff float64
    
    for scanner.Scan() {
        line := scanner.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := strings.Fields(line)
        
        if split[0] == "azimuth_pixel_offset" {
            diff, err = strconv.ParseFloat(split[1], 64)
            
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

//func (self IFG) AdaptFilt(opt AdaptFiltOpt) (ret IFG, cc Coherence, err error) {
    //step := float64(opt.FFTWindow) / 8.0
    
    //if opt.step > 0.0 {
        //step = opt.step
    //}
    
    //// TODO: figure out the name of the output files
    //ret, err = NewIFG(self.Dat + ".filt", "", "", "", "")
    
    //if err != nil {
        //err = Handle(err, "failed to create new interferogram struct")
        //return
    //}
    
    //cc, err = NewCoherence("", "")
    
    //if err != nil {
        //err = Handle(err, "failed to create new dataFile struct")
        //return
    //}
    
    ///*
    //if Empty(filt):
        //filt = 
    
    //if empty(cc is None:
        //cc = self.datfile + ".cc"
    //*/
    
    //rng := self.Rng
    
    //_, err = adf(self.Dat, ret.Dat, cc.Dat, rng, opt.alpha, opt.FFTWindow,
                 //opt.cohWindow, step, opt.offset.Azi, opt.offset.Rng,
                 //opt.frac)
    
    //if err != nil {
        //err = Handle(err, "adaptive filtering failed")
        //return
    //}
    
    //return ret, cc, nil
//}

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
