package data

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/command"
)

type Service interface {
    ComplexToReal(src, dst path.ValidFile, mode CpxToReal) error
}

type ServiceImpl struct {
    cpxToReal command.Command
}
