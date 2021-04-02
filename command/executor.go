package command

import (
    "os/exec"
)

type Executor interface {
    Execute(cmd Command, ctx Context) error
}

func Call(ex Executor, cmd Command, args []string) (err error) {
    return ex.Execute(cmd, DefaultContext.WithArgs(args))
}

type Execute struct {
}

var execute Execute

func NewExecute() (ex Execute) {
    return execute
}

func (e Execute) Execute(cmd Command, ctx Context) (err error) {
    return exec.CommandContext(ctx.Context, cmd.String(), ctx.Args...).Run()
}

type Setup struct {}

func (_ Setup) CreateExecutor() (e Executor, err error) {
    return NewExecute(), nil
}

func WithEnv(ex Executor, env Env) (se SharedEnv) {
    return SharedEnv {
        ex: ex,
        env: env,
    }
}
