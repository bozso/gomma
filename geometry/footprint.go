package geometry

import (
	"github.com/paulmach/orb"
)

type AreaOfInterest struct {
	Footprint orb.Geometry `json:"footprint"`
}
