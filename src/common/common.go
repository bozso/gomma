package common

import (
    "fmt"
    
    "../utils"
)

type RngAzi struct {
    Rng int `json:"rng"`
    Azi int `json:"azi"`
}

var DefRA = RngAzi{Rng:1, Azi:1}

func (ra RngAzi) String() string {
    return fmt.Sprintf("%d,%d", ra.Rng, ra.Azi)
}

func (ra *RngAzi) Set(s string) (err error) {
    var ferr = merr.Make("RngAzi.Decode")
    
    if len(s) == 0 {
        return ferr.Wrap(utils.EmptyStringError{})
    }
    
    split, err := utils.NewSplitParser(s, ",")
    if err != nil {
        return ferr.Wrap(err)
    }
    
    ra.Rng, err = split.Int(0)
    if err != nil {
        return ferr.Wrap(err)
    }

    ra.Azi, err = split.Int(1)
    if err != nil {
        return ferr.Wrap(err)
    }
    
    return nil
}

func (ra RngAzi) Check() (err error) {
    var ferr = merr.Make("RngAzi.Check")
     
    if ra.Rng == 0 {
        return ferr.Wrap(ZeroDimError{dim: "range samples / columns"})
    }
    
    if ra.Azi == 0 {
        return ferr.Wrap(ZeroDimError{dim: "azimuth lines / rows"})
    }
    
    return nil
}

func (ra *RngAzi) Default() {
    if ra.Rng == 0 {
        ra.Rng = 1
    }
    
    if ra.Azi == 0 {
        ra.Azi = 1
    }
}

