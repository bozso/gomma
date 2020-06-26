package data

import (
    "github.com/bozso/gotoolbox/path"
)

type Parameter struct {
    ParFile path.ValidFile `json:"parfile"`    
}

func (p *Parameter) SetParFile(vf path.ValidFile) {
    p.ParFile = vf
}

type FileWithPar struct {
    Parameter `json:"parameter"`
    File      `json:"file"`
}
