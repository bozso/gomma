package settings

import (
    "fmt"
    "log"
    
    "github.com/bozso/gotoolbox/command"
    "github.com/bozso/gotoolbox/errors"
)

type Command struct {
    command.Command
}

func NewCommand(name string) (c Command) {
    c.Command = command.New(name)
    return
}

type Commands map[string]Command

func (cs Commands) Get(name string) (c Command, err error) {
    c, ok := cs[name]
    if !ok {
        err = errors.KeyNotFound(name)
    }
    return
}

func (cs Commands) Select(names ...string) (c Command, err error) {
    var ok bool
    for _, name := range names {
        c, ok = cs[name]
        
        if ok {
            return
        }
    }
    
    err = fmt.Errorf("at least one executable from %s  must be an available executable",
        names)
    
    return
}

func (cs Commands) Must(name string) (c Command) {
    c, ok := cs[name]
    
    if !ok {
        log.Fatalf("failed to find Gamma executable '%s'", name)
    }
    
    return
}
