package batch

import (
	"log"

	"github.com/bozso/gomma/cli"
	"github.com/bozso/gomma/settings"
	"github.com/bozso/gotoolbox/path"
)

type ExecutorPayload struct {
	Tag  string `json:"type"`
	Data []byte `json:"data"`
}

type ContextPayload struct {
	Logfile   path.File         `json:"logfile"`
	GammaPath path.Dir          `json:"gamma_dir"`
	Executor  ExecutorPayload   `json:"executor"`
	Config    cli.PathOrPayload `json:"config"`
}

type Context struct {
	Logger        *log.Logger
	GammaCommands settings.Commands
	config        cli.PathOrPayload
	/// TODO: add executor after merge
}

func (c *Context) Set(s string) (err error) {
	decoder := cli.CoderJSON.GetDecoder()

	f, err := path.New(s).ToValidFile()
	if err != nil {
		return
	}

	reader, err := f.Open()
	if err != nil {
		return
	}
	defer reader.Close()

	var payload ContextPayload
	err = decoder.Decode(&payload)
	if err != nil {
		return
	}

	c.config = payload.Config

	return nil
}
