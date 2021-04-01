package command

import (
    "fmt"
    "io"
    "strings"
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

func WithEnv(ex Executor, env Env) (se SharedEnv) {
    return SharedEnv {
        ex: ex,
        env: env,
    }
}


