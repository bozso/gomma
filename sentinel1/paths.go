package sentinel1

import (
	"fmt"

	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/data"
	"github.com/bozso/gomma/utils/params"
)

type IWPath struct {
	data.PathWithPar
	TOPSPar path.File
}

type IWPaths [maxIW]IWPath

func NewIW(datafile path.File) (p IWPath) {
	p.DatFile = datafile
	p.ParFile = datafile.AddExt("par").ToFile()
	p.TOPSPar = datafile.AddExt("TOPS_par").ToFile()

	return
}

func (p IWPath) WithPar(par path.File) (pp IWPath) {
	p.ParFile = par
	return p
}

func (p IWPath) WithTOPS(tops path.File) (pp IWPath) {
	p.TOPSPar = tops
	return p
}

func (iw IWPath) Tabline() (s string) {
	s = fmt.Sprintf("%s %s %s\n", iw.DatFile, iw.ParFile, iw.TOPSPar)
	return
}

func (p IWPath) GetParser() (pp params.Parser, err error) {
	p1, err := p.GetParser()
	if err != nil {
		return
	}

	par, err := p.TOPSPar.ToValid()

	p2, err := data.NewGammaParams(par)
	if err != nil {
		return
	}

	pp = params.NewTeeParser(p1, p2).ToParser()
	return
}

func (p IWPath) Load() (iw IW, err error) {
	pp, err := p.GetParser()
	if err != nil {
		return
	}

	iw.TOPSPar, err = p.TOPSPar.ToValid()
	if err != nil {
		return
	}

	err = p.WithParser(pp).Load(&iw)

	return
}

type SLCPath struct {
	Tab path.File
	SLCMeta
	IWPaths
}

func (sp SLCPath) CreateTabfile() (err error) {
	file, err := sp.Tab.Create()
	if err != nil {
		return
	}
	defer file.Close()

	for ii := 0; ii < sp.nIW; ii++ {
		_, err = file.WriteString(sp.IWPaths[ii].Tabline())
		if err != nil {
			return
		}
	}
	return
}

func (sp SLCPath) Load() (slc SLC, err error) {
	for ii := 0; ii < sp.nIW; ii++ {
		iw := &sp.IWPaths[ii]

		slc.IWs[ii], err = iw.Load()
		if err != nil {
			return
		}
	}

	slc.Tab, err = sp.Tab.ToValid()
	if err != nil {
		return
	}

	slc.SLCMeta = sp.SLCMeta
	return
}
