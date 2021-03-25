package command

import (
    "github.com/bozso/gotoolbox/path"
)

type Command struct {
    binPath path.ValidFile
}

func (c Command) String() (s string) {
    return c.binPath.String()
}
