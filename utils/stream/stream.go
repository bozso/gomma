package stream

import (
    "fmt"
)

type Stream struct {
    name string
}

func (s Stream) String() string {
    return s.name
}

func (s Stream) Fail(op string, err error) error {
    return StreamError{s.name, op, err}
}

func (s Stream) OpenFail(err error) error {
    return s.Fail("open", err)
}

func (s Stream) CreateFail(err error) error {
    return s.Fail("create", err)
}

type StreamError struct {
    path, op string
    err error
}

func (e StreamError) Error() string {
    return fmt.Sprintf("failed to %s from '%s'", e.op, e.path)
}

func (e StreamError) Unwrap() error {
    return e.err
}

