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

type SharedEnv struct {
    env Env
    ex Executor
}

func (se SharedEnv) Execute(cmd Command, ctx Context) (err error) {
    ctx.Env = se.env
    return se.ex.Execute(cmd, ctx)
}

func Format(cmd Command, ctx Context) (s string) {
    return fmt.Sprintf("%s %s %s",
        strings.Join(ctx.Env.Get(), " "),
        cmd.String(),
        strings.Join(ctx.Args, " "),
    )
}

type Debugger struct {
    wr io.Writer
}

func (d Debugger) Execute(cmd Command, ctx Context) (err error) {
    _, err = fmt.Fprintf(d.wr, "command to be executed\n%s\n", Format(cmd, ctx))
    return
}
