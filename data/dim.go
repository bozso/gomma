package data

import (
	"fmt"
)

type RngAzi struct {
	Rng uint64 `json:"range"`
	Azi uint64 `json:"azimuth"`
}

type ShapeMismatchError struct {
	Along    string
	One, Two uint64
}

func (s ShapeMismatchError) Error() string {
	return fmt.Sprintf(
		"expected dimension '%s' to have the same extent (%d != %d)", s.Along, s.One, s.Two)
}

func (r RngAzi) SameCols(other RngAzi) (b bool) {
	return r.Rng == other.Rng
}

func (r RngAzi) MustSameCols(other RngAzi) (err error) {
	if !r.SameCols(other) {
		return &ShapeMismatchError{
			One:   r.Rng,
			Two:   other.Rng,
			Along: "range samples / columns",
		}
	}

	return nil
}

func (r RngAzi) SameRows(other RngAzi) (b bool) {
	return r.Azi == other.Azi
}

func (r RngAzi) MustSameRows(other RngAzi) (err error) {
	if !r.SameRows(other) {
		return &ShapeMismatchError{
			One:   r.Azi,
			Two:   other.Azi,
			Along: "azimuth lines / rows",
		}
	}

	return nil
}

func (r RngAzi) SameShape(other RngAzi) (b bool) {
	b = r.SameCols(other)
	if !b {
		return false
	}

	return r.SameRows(other)
}

type DimMismatchError struct {
	One, Two RngAzi
}

func (d DimMismatchError) Error() (s string) {
	return fmt.Sprintf("expected dimensions to match (%v != %v)",
		d.One, d.Two)
}

func (r RngAzi) MustSameShape(other RngAzi) (err error) {
	if !r.SameShape(other) {
		return &DimMismatchError{
			One: r,
			Two: other,
		}
	}

	return nil
}
