package command

import (
    "io"
    "fmt"

    "github.com/bozso/gomma/stream"
)


type Debugger struct {
    wr io.Writer
}

func (d Debugger) Execute(cmd Command, ctx Context) (err error) {
    _, err = fmt.Fprintf(d.wr, "command to be executed\n%s\n", Format(cmd, ctx))
    return
}

type DebuggerConfig struct {
    logfile stream.Config `json:"logfile"`
}

func (d *DebuggerConfig) ToExecutor() (e Executor, err error) {
    wr, err := d.logfile.ToOutStream()
    if err != nil {
        return
    }

    e = Debugger {
        wr: &wr,
    }
    return
}
