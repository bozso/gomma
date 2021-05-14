package service

import (
	"github.com/bozso/gomma/settings"
)

type Setup interface {
	SetupGamma(settings.Setup) error
	Get(string) (settings.Command, error)
}

type Sentinel1 interface {
	SelectFiles(SentinelSelect) error
	DataImport(SentinelImport) error
}
