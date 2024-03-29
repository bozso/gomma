package sentinel1

import (
	"fmt"
	"path/filepath"

	"github.com/bozso/gomma/common"
	"github.com/bozso/gotoolbox/path"
)

type tplType int

const (
	tiff tplType = iota
	annot
	calib
	noise
	preview
	quicklook
)

var fmtNeeded = [nTemplate]bool{
	tiff:      true,
	annot:     true,
	calib:     true,
	noise:     true,
	preview:   false,
	quicklook: false,
}

var calibPath = filepath.Join("annotation", "calibration")

const nTemplate = 6

type Template interface {
	Render(ii int, pol common.Pol) string
}

type templates [nTemplate]Template

type noFormat struct {
	tpl string
}

func (n noFormat) Render(_ int, _ common.Pol) string {
	return n.tpl
}

type format struct {
	tpl string
}

func (f format) Render(ii int, pol common.Pol) string {
	return fmt.Sprintf(f.tpl, ii, pol)
}

func newTemplates(safe path.File, tpl string) templates {
	return templates{
		tiff: format{
			safe.Join("measurement", tpl+".tiff").GetPath(),
		},

		annot: format{
			safe.Join("annotation", tpl+".xml").GetPath(),
		},

		calib: format{
			safe.Join(calibPath,
				fmt.Sprintf("calibration-%s.xml", tpl)).GetPath(),
		},

		noise: format{
			safe.Join(calibPath, fmt.Sprintf("noise-%s.xml", tpl)).GetPath(),
		},

		preview: noFormat{
			safe.Join("preview", "product-preview.html").GetPath(),
		},

		quicklook: noFormat{
			safe.Join("preview", "quick-look.png").GetPath(),
		},
	}
}
