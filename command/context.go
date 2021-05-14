package command

import (
	"context"
)

var DefaultContext = Context{
	Env:     EmptyEnv,
	Context: context.Background(),
	Args:    emptySlice,
}

type Context struct {
	Args    []string
	Env     Env
	Context context.Context
}

func (c Context) WithArgs(args []string) (out Context) {
	out = c
	out.Args = args
	return out
}
