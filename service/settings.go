package service

import (
	"sync"

	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/settings"
)

type Settings struct {
	mutex    sync.RWMutex
	settings settings.Settings
	commands settings.Commands
}

type Directory struct {
	Dir path.Dir `json:"directory"`
}

func (s *Settings) SetupGamma(ss settings.Setup) (err error) {
	s.settings, err = ss.New()
	if err != nil {
		return
	}

	s.commands, err = s.settings.MakeCommands()
	return
}

func (s *Settings) Get(name string) (c settings.Command, err error) {
	s.mutex.RLock()
	c, err = s.commands.Get(name)
	s.mutex.RUnlock()
	return
}
