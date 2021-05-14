package settings

import (
	"fmt"
	"log"

	"github.com/bozso/gomma/command"
	"github.com/bozso/gotoolbox/errors"
)

type Commands map[string]command.Command

func (cs Commands) Get(name string) (c command.Command, err error) {
	c, ok := cs[name]
	if !ok {
		err = errors.KeyNotFound(name)
	}
	return
}

func (cs Commands) Select(names ...string) (c command.Command, err error) {
	var ok bool
	for _, name := range names {
		c, ok = cs[name]

		if ok {
			return
		}
	}

	err = fmt.Errorf("at least one executable from %s must be an available executable",
		names)

	return
}

func (cs Commands) Must(name string) (c command.Command) {
	c, ok := cs[name]

	if !ok {
		log.Fatalf("failed to find Gamma executable '%s'", name)
	}

	return
}
