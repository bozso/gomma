package datafile

import (
    "fmt"
    
    "../utils"
)

const (
    TimeParseErr utils.Werror = "failed retreive %s from date string '%s'"
    RngError utils.Werror = "failed to retreive range samples from '%s'"
    AziError utils.Werror = "failed to retreive azimuth lines from '%s'"
)


type ParameterError struct {
    path, par string
    err error
}

func (p ParameterError) Error() string {
    return fmt.Sprintf("failed to retreive parameter '%s' from file '%s'",
        p.par, p.path)
}

func (p ParameterError) Unwrap() error {
    return p.err
}

type TypeMismatchError struct {
    ftype, expected string
    DType
    Err error
}

func (e TypeMismatchError) Error() string {
    return fmt.Sprintf("expected datatype '%s' for %s datafile, got '%s'",
        e.expected, e.ftype, e.DType.String())
}

func (e TypeMismatchError) Unwrap() error {
    return e.Err
}

type UnknownTypeError struct {
    DType
    Err error
}

func (e UnknownTypeError) Error() string {
    return fmt.Sprintf("unrecognised type '%s', expected a valid datatype",
        e.DType.String())
}

func (e UnknownTypeError) Unwrap() error {
    return e.Err
}

type WrongTypeError struct {
    DType
    kind string
    Err error
}

func (e WrongTypeError) Error() string {
    return fmt.Sprintf("wrong datatype '%s' for %s", e.kind, e.DType.String())
}

func (e WrongTypeError) Unwrap() error {
    return e.Err
}

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

type ShapeMismatchError struct {
    dat1, dat2, dim string
    n1, n2 int
}

func (s ShapeMismatchError) Error() string {
    return fmt.Sprintf("expected datafile '%s' to have the same %s as " + 
                       "datafile '%s' (%d != %d)", s.dat1, s.dim, s.dat2, s.n1,
                       s.n2)
}
