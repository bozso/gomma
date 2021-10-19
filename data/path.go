package data

type Path struct {
	DataFile string
}

func (p Path) UnmarshalJSON(b []byte) (err error) {
	p.DataFile = string(b)
	return nil
}

func (p Path) MarshalJSON() (b []byte, err error) {
	return []byte(p.DataFile), nil
}

func New(file string) (p Path) {
	p.DataFile = file

	return
}

func (p Path) WithParFile(file string) (pp PathWithPar) {
	return PathWithPar{
		Path:    p,
		ParFile: file,
	}
}

type PathWithPar struct {
	Path    Path
	ParFile string
}

func (p PathWithPar) WithPar(file string) (pp PathWithPar) {
	p.ParFile = file

	return p
}
