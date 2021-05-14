package data

import (
	"github.com/bozso/gotoolbox/command"
	"github.com/bozso/gotoolbox/path"
)

type Service interface {
	ComplexToReal(src, dst path.ValidFile, mode CpxToReal) error
}

type ServiceImpl struct {
	cpxToReal command.Command
}
