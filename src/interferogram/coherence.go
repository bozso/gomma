package interferogram

import (
    "../data"
)

type CoherenceWeight int

const (
    Constant CoherenceWeight = iota
    Gaussian
)

type (
    Coherence struct {
        data.File
    }
    
    CoherenceOpt struct {
        Box                    common.Minmax
        SlopeCorrelationThresh float64
        SlopeWindow            int
        DatFile,  string
        Weight CoherenceWeight
    }
)



var (
    phaseSlope = Gamma.Must("phase_slope")
    CCAdaptive = Gamma.Must("cc_ad")
)

func (ifg File) Coherence(opt CoherenceOpt) (c Coherence, err error) {
    weightFlag := CoherenceWeight[opt.WeightType]
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width := ifg.Rng
    
    log.Printf("Estimating phase slope.")
    
    // TODO: figure out name
    slope := ".cpx"
    
    // parameters: xmin, xmax, ymin, ymax not yet given
    _, err = phaseSlope(ifg.Dat, slope, opt.SlopeWindow,
                        opt.SlopeCorrelationThresh)
    
    if err != nil {
        return
    }

    log.Printf("Calculating coherence.")
    
    mli1, mli2 := "", ""
    
    _, err = CCAdaptive(ifg.Dat, mli1, mli2, slope, nil, c.Dat, width,
                        opt.Box.Min, opt.Box.Max, weightFlag)
    
    return
}
