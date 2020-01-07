package gamma

import (
    "time"
    "log"
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

func NewIFG(dat, off, diffpar string) (ifg IFG, err error) {
    var ferr = merr.Make("NewIFG")

    if ifg.DatParFile, err = NewDatParFile(dat, off, "off", FloatCpx);
       err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if len(diffpar) == 0 {
        diffpar = dat + ".diff_par"
    }
    
    ifg.DiffPar = NewGammaParam(diffpar)
    
    return
}

func TmpIFG() (ret IFG, err error) {
    if ret.DatParFile, err = TmpDatParFile("diff", "off", FloatCpx); err != nil {
        return
    }
    
    ret.DiffPar = Params{Par: ret.Dat + ".par", Sep: ":"}
    ret.SimUnwrap = ret.Dat + ".sim_unw"
    
    return ret, nil
}

func (ifg IFG) jsonMap() (js JSONMap) {
    js = ifg.DatParFile.jsonMap()
    
    js["quality"] = ifg.Quality
    js["diffparfile"] = ifg.DiffPar
    js["simulated_unwrapped"] = ifg.SimUnwrap
    
    return
}

func (i IFG) Move(dir string) (im IFG, err error) {
    var ferr = merr.Make("IFG.Move")
    
    if im.DatParFile, err = i.DatParFile.Move(dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    if im.DiffPar.Par, err = Move(i.DiffPar.Par, dir); err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    im.DiffPar.Sep = ":"
    
    if len(i.SimUnwrap) > 0 {
        if im.SimUnwrap, err = Move(i.SimUnwrap, dir); err != nil {
            err = ferr.Wrap(err)
            return
        }
    }
    
    if len(i.Quality) > 0 {
        if im.Quality, err = Move(i.Quality, dir); err != nil {
            err = ferr.Wrap(err)
            return
        }
    }
    
    return
}

func (i *IFG) FromJson(m JSONMap) (err error) {
    var ferr = merr.Make("IFG.FromJson")
    
    if err = i.DatParFile.FromJson(m); err != nil {
        return ferr.Wrap(err)
    }
    
    if i.DType != FloatCpx {
        err = TypeMismatchError{ftype:"IFG", expected:"complex",
            DType:i.DType}
        return ferr.Wrap(err)
    }
    
    if i.Quality, err = m.String("quality"); err != nil {
        err = ferr.WrapFmt(err, "failed to retreive quality file")
        return
    }
    
    if i.DiffPar.Par, err = m.String("diffparfile"); err != nil {
        err = ferr.WrapFmt(err, "failed to diffparfile")
        return
    }
    i.DiffPar.Sep = ":"
    
    
    if i.SimUnwrap, err = m.String("simulated_unwrapped"); err != nil {
        err = ferr.WrapFmt(err, "failed to simulated unwrapped datafile")
        return
    }
    
    return nil
}
    
func FromSLC(slc1, slc2, ref *SLC, opt IfgOpt) (ifg IFG, err error) {
    var ferr = merr.Make("FromSLC")
    inter := 0
    
    if opt.interact {
        inter = 1
    }
    
    rng, azi := opt.Looks.Rng, opt.Looks.Azi
    
    par1, par2 := slc1.Par, slc2.Par
    
    // TODO: check arguments!
    _, err = createOffset(par1, par2, ifg.Par, opt.algo, rng, azi, inter)
    
    if err != nil {
        err = ferr.WrapFmt(err, "failed to create offset table")
        return
    }
    
    slcRefPar := "-"
    
    if ref != nil {
        slcRefPar = ref.Par
    }
    
    if ifg, err = TmpIFG(); err != nil {
        err = ferr.Wrap(err)
        return 
    }
    
    _, err = phaseSimOrb(par1, par2, ifg.Par, opt.hgt, ifg.SimUnwrap,
        slcRefPar, nil, nil, 1)
    
    if err != nil {
        err = ferr.Wrap(err)
        return 
    }

    dat1, dat2 := slc1.Dat, slc2.Dat
    _, err = slcDiffIntf(dat1, dat2, par1, par2, ifg.Par,
        ifg.SimUnwrap, ifg.DiffPar, rng, azi, 0, 0)
    
    if err != nil {
        err = ferr.Wrap(err)
        return 
    }
    
    if err = ifg.Parse(); err != nil {
        err = ferr.Wrap(err)
        return 
    }
    
    // TODO: Check date difference order
    ifg.DeltaT = slc1.Time.Sub(slc2.Time)
    
    return
}

const (
    Real CpxToReal = iota
    Imaginary
    Intensity
    Magnitude
    Phase
)

func (c CpxToReal) String() string {
    switch c {
    case Real:
        return "Real"
    case Imaginary:
        return "Imaginary"
    case Intensity:
        return "Intensity"
    case Magnitude:
        return "Magnitude"
    case Phase:
        return "Phase"
    default:
        return "Unknown"
    }
}

var cpxToReal = Gamma.Must("cpx_to_real")

func (ifg IFG) ToReal(mode CpxToReal, name string) (d DatFile, err error) {
    var ferr = merr.Make("IFG.ToReal")
    
    if len(name) == 0 {
        d, err = TmpDatFile("real", Float)
    } else {
        d, err = NewDatFile(name, Float)
    }
    
    if err != nil {
        err = ferr.Wrap(err)
        return
    }
    
    d.URngAzi = ifg.URngAzi
    
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
        err = ferr.Wrap(ModeError{name:"IFG.ToReal", got:mode})
        return
    }
    
    if _, err = cpxToReal(ifg.Dat, d.Dat, d.Rng, Mode); err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

var rasmph_pwr24 = Gamma.Must("rasmph_pwr24")

func (ifg IFG) Raster(opt RasArgs) (err error) {
    var ferr = merr.Make("IFG.Raster")

    opt.Mode = MagPhasePwr
    
    if err = ifg.Raster(opt); err != nil {
        err = ferr.Wrap(err)
    }
    return
}

func (ifg IFG) rng() (i int, err error) {
    var ferr = merr.Make("IFG.rng")
    
    if i, err = ifg.Int("interferogram_width", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return 
}

func (ifg IFG) azi() (i int, err error) {
    var ferr = merr.Make("IFG.azi")
    
    if i, err = ifg.Int("interferogram_azimuth_lines", 0); err != nil {
        err = ferr.Wrap(err)
    }
    
    return 
}

/*
 * TODO: remove?
func (ifg IFG) imgfmt() (string, error) {
    return "FCOMPLEX", nil
}
*/


func (ifg IFG) CheckQuality() (b bool, err error) {
    var (
        ferr = merr.Make("IFG.CheckQuality")
        qual = ifg.Quality
    )
    
    var file Reader
    if file, err = NewReader(qual); err != nil {
        err = ferr.Wrap(err)
        return
    }
    defer file.Close()
    
    offs := 0.0
    var diff float64
    for file.Scan() {
        line := file.Text()
        
        if len(line) == 0 {
            continue
        }
        
        split := strings.Fields(line)
        
        if split[0] == "azimuth_pixel_offset" {
            s := split[1]
            diff, err = strconv.ParseFloat(s, 64)
            
            if err != nil {
                err = ferr.WrapFmt(err,
                    "failed to parse: '%s' into float64", s)
                return
            }
            
            offs += diff
        }
    }
    
    log.Printf("Sum of azimuth offsets in %s is %f pixel.\n", qual, offs)
    
    if offs > 0.0 || offs < 0.0 {
        b = true
    } else {
        b = false
    }
    
    return
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

func (ifg IFG) Coherence(opt CoherenceOpt) (c Coherence, err error) {
    var ferr = merr.Make("IFG.Coherence")
    weightFlag := CoherenceWeight[opt.WeightType]
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width := ifg.Rng
    
    log.Printf("Estimating phase slope. ")
    
    // TODO: figure out name
    slope := ".cpx"
    
    // parameters: xmin, xmax, ymin, ymax not yet given
    _, err = phaseSlope(ifg.Dat, slope, opt.SlopeWindow,
                        opt.SlopeCorrelationThresh)
    
    if err != nil {
        err = ferr.Wrap(err)
        return
    }

    log.Printf("Calculating coherence. ")
    
    mli1, mli2 := "", ""
    
    _, err = CCAdaptive(ifg.Dat, mli1, mli2, slope, nil, c.Dat, width,
                        opt.Box.Min, opt.Box.Max, weightFlag)
    
    if err != nil {
        err = ferr.Wrap(err)
    }
    
    return
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
