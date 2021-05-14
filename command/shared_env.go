package command

import ()

type SharedEnv struct {
	env Env
	ex  Executor
}

func (se SharedEnv) Execute(cmd Command, ctx Context) (err error) {
	ctx.Env = se.env
	return se.ex.Execute(cmd, ctx)
}
