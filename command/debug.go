package command

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bozso/gomma/stream"
)

type Formatter interface {
	FormatCommand(io.Writer, Command, Context) (int, error)
}

type Debug struct {
	wr  io.Writer
	fmt Formatter
}

func (d *Debug) UnmarshalJSON(b []byte) (err error) {
	var out stream.Out
	if err = json.Unmarshal(b, &out); err != nil {
		return
	}
	d.wr = &out
	d.fmt = LineFormat
	return nil
}

func (d Debug) Execute(cmd Command, ctx Context) (err error) {
	_, err = d.fmt.FormatCommand(d.wr, cmd, ctx)
	return
}

/*
type DebugConfig struct {
	Logfile stream.Config `json:"logfile"`
}

func (d DebugConfig) CreateExecutor() (e Executor, err error) {
	wr, err := d.Logfile.ToOutStream()
	if err != nil {
		return
	}

	e = Debug{
		wr:  &wr,
		fmt: LineFormat,
	}
	return
}
*/

var LineFormat Formatter = (*LineFormatter)(nil)

type LineFormatter struct{}

func (_ *LineFormatter) FormatCommand(wr io.Writer, cmd Command, ctx Context) (n int, err error) {
	return fmt.Fprintf(wr, "%s %s %s",
		strings.Join(ctx.Env.Get(), " "),
		cmd.String(),
		strings.Join(ctx.Args, " "),
	)
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

type CreateJSONEncoder struct{}

func (_ CreateJSONEncoder) CreateEncoder(wr io.Writer) (e Encoder) {
	return json.NewEncoder(wr)
}

func (e EncodeFormatter) FormatCommand(wr io.Writer, cmd Command, ctx Context) (n int, err error) {
	type pair struct {
		Command Command
		Context Context
	}

	err = e.creator.CreateEncoder(wr).Encode(pair{cmd, ctx})
	return
}
