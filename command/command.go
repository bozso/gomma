package command

import (
	"github.com/bozso/gotoolbox/path"
)

type Command struct {
	binPath path.ValidFile
}

func New(p path.Path) (c Command, err error) {
	c.binPath, err = p.ToValidFile()

	return
}

func (c Command) String() (s string) {
	return c.binPath.String()
}
