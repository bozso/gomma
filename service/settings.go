package service

import (
    "sync"
    
    "github.com/bozso/gotoolbox/path"
    
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

func (s *Settings) Setup(s settings.Setup) (err error) {
    s.settings, err = s.New()
    if err != nil {
        return 
    }
    
    s.commands, err = s.settings.MakeCommands()
    return
}

func (s *Settings) Get(name string) (c command.Command, err error) {
    s.mutex.RLock()
    c, err = s.commands.Get(name)
    s.mutex.RUnlock()
    return
}
