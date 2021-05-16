package command

import (
	"fmt"
	"io"
	"strings"

	json "git.sr.ht/~istvan_bozso/sert/json"

	"github.com/bozso/gomma/stream"
)

type Formatter interface {
	FormatCommand(io.Writer, Command, Context) (int, error)
}

type Debug struct {
	Out *stream.Out   `json:"logfile"`
	Fmt FormatterJSON `json:"formatter"`
}

func (d Debug) Execute(cmd Command, ctx Context) (err error) {
	_, err = d.Fmt.FormatCommand(d.Out, cmd, ctx)
	return
}

type FormatterJSON struct {
	Formatter
}

func (f *FormatterJSON) UnmarshalJSON(b []byte) (err error) {
	switch json.Trim(b) {
	case "line_formatter":
		f.Formatter = LineFormat
	default:
		err = fmt.Errorf("unrecognized option for formatter: '%s'", b)
	}

	return
}

var LineFormat Formatter = (*LineFormatter)(nil)

type LineFormatter struct{}

func (_ *LineFormatter) FormatCommand(wr io.Writer, cmd Command, ctx Context) (n int, err error) {
	return fmt.Fprintf(wr, "%s %s %s",
		strings.Join(ctx.Env.Get(), " "),
		cmd.String(),
		strings.Join(ctx.Args, " "),
	)
}
