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

/*

func (d File) WithShape(p Path) (f File, err error) {
	f, err = d.WithShapeDType(p, d.Dtype)
	return
}

func (d File) WithShapeDType(p Path, dtype Type) (f File, err error) {
	if dtype == Unknown {
		dtype = d.Dtype
	}

	f, err = p.Load(d.RngAzi, dtype)
	return
}

func (p Path) Load(ra common.RngAzi, dtype Type) (f File, err error) {
	f.DataFile, err = p.DataFile.ToValidFile()
	if err != nil {
		return
	}

	f.RngAzi, f.Dtype = ra, dtype
	return
}
*/
