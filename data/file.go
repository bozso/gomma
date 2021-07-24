package data

import ()

type File struct {
	DataFile Path
	Meta     Meta `json:"meta"`
}

func (f File) Rng() (rng uint) {
	return f.Meta.RngAzi.Rng
}

func (f File) Azi() (azi uint) {
	return f.Meta.RngAzi.Azi
}

func (f File) SameDim(other File) (b bool) {
	return f.Meta.RngAzi.SameShape(other.Meta.RngAzi)
}

func (f File) MustSameDim(other File) (err error) {
	return f.Meta.RngAzi.MustSameShape(other.Meta.RngAzi)
}

func (f File) AsFile() (F File) {
	return f
}

type Like interface {
	AsFile() File
}
