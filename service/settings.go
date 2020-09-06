package service

import (
    "sync"
    
    "github.com/bozso/gomma/command"
    "github.com/bozso/gomma/settings"
)

type Settings struct {
    mutex    sync.RWMutex
    settings settings.Settings
    commands command.Commands
}

type Directory struct {
    Dir path.Dir `json:"directory"`
}

//func (s *Settings) SetGammaDirectory()

func (s *Settings) Get(name string) (c command.Command, err error) {
    s.mutex.RLock()
    c, err s.commands.Get(name)
    s.mutex.RUnlock()
    return
}
