package plot

import (
    "github.com/bozso/gamma/common"
)

type ScaleExp struct {
    Scale float64 `name:"scale" default:"1.0"`
    Exp   float64 `name:"exp" default:"0.35"`
}

func (se *ScaleExp) Set(s string) (err error) {
    // implement!
    return
}

func (se *ScaleExp) Parse() {
    if se.Scale == 0.0 {
        se.Scale = 1.0
    }
    
    if se.Exp == 0.0 {
        se.Exp = 0.35
    }
}

type ZeroFlag int

const (
    Missing ZeroFlag = iota
    Valid
)

type Inverse int

const (
    Float2Raster Inverse = 1
    Raster2Float Inverse = -1
)

type Channel int

const (
    Red   Channel = 1
    Green Channel = 2
    Blue  Channel = 3
)

func calcFactor(ndata, factor int) int {
    // log.Printf("ndata: %d factor: %d\n", ndata, factor)
    
    ret := float64(ndata) / float64(factor)
    
    // log.Fatalf("ret: %f\n", ret)
    
    if ret <= 0.0 {
        return 1
    } else {
        return int(ret)
    }
}



var (
    rasByte = common.Gamma.Must("rasbyte")
    rasCC = common.Gamma.Must("rascc")
    rasdB = common.Gamma.Must("ras_dB")
    rasHgt = common.Gamma.Must("rashgt")
    rasdtPwr = common.Gamma.Must("rasdt_pwr")
    rasMph = common.Gamma.Must("rasmph")
    rasMphPwr = common.Gamma.Must("rasmph_pwr")
    rasPwr = common.Gamma.Must("raspwr")
    rasRmg = common.Gamma.Must("rasrmg")
    rasShd = common.Gamma.Must("rasshd")
    rasSLC = common.Gamma.Must("rasSLC")
    rasLinear = common.Gamma.Must("ras_linear")
)