package common

import (
    "fmt"
    
    "../utils"
)

const merr utils.ModuleName = "gamma.common"

type ZeroDimError struct {
    dim string
    Err error
}

func (e ZeroDimError) Error() string {
    return fmt.Sprintf("expected %s to be non zero", e.dim)
}

func (e ZeroDimError) Unwrap() error {
    return e.Err
}

type DateParseError struct {
    source string
    err error
}

func (d DateParseError) Error() string {
    return fmt.Sprintf("failed to parse date from string: '%s'", d.source)
}

func (d DateParseError) Unwrap() error {
    return d.err
} 
