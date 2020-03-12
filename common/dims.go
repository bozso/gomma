package common

import (
    "fmt"
    "github.com/bozso/gotoolbox/splitted"
    "github.com/bozso/gotoolbox/errors"
)

type RngAzi struct {
    Rng int `json:"rng"`
    Azi int `json:"azi"`
}

var DefaultRngAzi = RngAzi{Rng:1, Azi:1}

func (ra RngAzi) String() string {
    return fmt.Sprintf("%d,%d", ra.Rng, ra.Azi)
}

func (ra *RngAzi) Set(s string) (err error) {
    
    if err = errors.NotEmpty(s, "RngAzi"); err != nil {
        return
    }
    
    split, err := splitted.New(s, ",")
    if err != nil { return }
    
    ra.Rng, err = split.Int(0)
    if err != nil { return }

    ra.Azi, err = split.Int(1)
    
    return
}

func (ra *RngAzi) Default() {
    if ra.Rng == 0 {
        ra.Rng = 1
    }
    
    if ra.Azi == 0 {
        ra.Azi = 1
    }
}

func (ra RngAzi) Validate() (err error) {
    if ra.Rng == 0 {
        return ZeroDimError{dim: "range samples / columns"}
    }
    
    if ra.Azi == 0 {
        return ZeroDimError{dim: "azimuth lines / rows"}
    }
    
    return nil
}

type Dims interface {
    Rng() int
    Azi() int
}

type ZeroDimError struct {
    dim string
    err error
}

func (e ZeroDimError) Error() string {
    return fmt.Sprintf("expected %s to be non zero", e.dim)
}

func (e ZeroDimError) Unwrap() error {
    return e.err
}
