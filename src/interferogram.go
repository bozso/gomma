package gamma

import (
    "time"
    "log"
    "strings"
    "strconv"
    
    "./data"
)

type (
    CohWeight int
    OffsetAlgo int
    CpxToReal  int
    
    Coherence struct {
        DatFile `json:"DatFile"`
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
