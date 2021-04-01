package command

import (
    "io"
    "fmt"
    "strings"
    "encoding/json"

    "github.com/bozso/gomma/stream"
)

type Formatter interface {
    FormatCommand(io.Writer, Command, Context) (int, error)
}

type Debugger struct {
    wr io.Writer
    fmt Formatter
}

func (d Debugger) Execute(cmd Command, ctx Context) (err error) {
    _, err = d.fmt.FormatCommand(d.wr, cmd, ctx)
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
        fmt: LineFormat,
    }
    return
}

var LineFormat Formatter = LineFormatter{}

type LineFormatter struct {}

func (_ LineFormatter) FormatCommand(wr io.Writer, cmd Command, ctx Context) (n int, err error) {
    return fmt.Fprintf(wr, "%s %s %s",
        strings.Join(ctx.Env.Get(), " "),
        cmd.String(),
        strings.Join(ctx.Args, " "),
    )
}

type Pair struct {
    Command Command
    Context Context
}

type Encoder interface {
    Encode(v interface{}) error
}

type EncoderCreator interface {
    CreateEncoder(io.Writer) Encoder
}

type EncodeFormatter struct {
    creator EncoderCreator
}

var JSONEncoderCreator EncoderCreator = CreateJSONEncoder{}

type CreateJSONEncoder struct {}

func (_ CreateJSONEncoder) CreateEncoder(wr io.Writer) (e Encoder) {
    return json.NewEncoder(wr)
}

func (e EncodeFormatter) FormatCommand(wr io.Writer, cmd Command, ctx Context) (n int, err error) {
    err = e.creator.CreateEncoder(wr).Encode(Pair{cmd, ctx})
    return
}
