package interferogram

import (
    "log"
    
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/data"
)

type CoherenceWeight int

const (
    Constant CoherenceWeight = iota
    Gaussian
)

func (cw CoherenceWeight) String() string {
    switch cw {
    case Constant:
        return "constant"
    case Gaussian:
        return "gaussian"
    default:
        return "unknown"
    }
}

type (
    Coherence struct {
        data.FloatFile
    }
    
    CoherenceOpt struct {
        Box                    common.Minmax
        SlopeCorrelationThresh float64
        SlopeWindow            int
        DatFile string
        Weight CoherenceWeight
    }
)


var (
    phaseSlope = common.Must("phase_slope")
    ccAdaptive = common.Must("cc_ad")
)

func (ifg File) Coherence(opt CoherenceOpt, c Coherence) (err error) {
    weightFlag := 0
    
    switch w := opt.Weight; w {
    case Constant:
        weightFlag = 0
    case Gaussian:
        weightFlag = 1
    default:
        return utils.UnrecognizedMode(w.String(),
            "adaptive coherence calculation")
    }
    
    //log.info("CALCULATING COHERENCE AND CREATING QUICKLOOK IMAGES.")
    //log.info('Weight type is "%s"'.format(weight_type))
    
    width := ifg.Ra.Rng
    
    log.Printf("Estimating phase slope.")
    
    // TODO: figure out a name
    slope := ".cpx"
    
    // parameters: xmin, xmax, ymin, ymax not yet given
    _, err = phaseSlope.Call(ifg.DatFile, slope, opt.SlopeWindow,
                        opt.SlopeCorrelationThresh)
    
    if err != nil { return }

    log.Printf("Calculating coherence.")
    
    mli1, mli2 := "", ""
    
    _, err = ccAdaptive.Call(ifg.DatFile, mli1, mli2, slope, nil,
        c.DatFile, width, opt.Box.Min, opt.Box.Max, weightFlag)
    
    return
}
