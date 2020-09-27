package data

import (
    
)

type Real struct {
    File `json:"file"`    
}

func (r Real) Validate() (err error) {
    return r.EnsureFloat()
}

type RealWithPar struct {
    Real
    Parameter `json:"parameter"`
}
