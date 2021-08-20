package data

type Path struct {
	DataFile string `json:"data_file"`
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
