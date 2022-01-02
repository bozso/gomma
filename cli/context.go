package cli

import (
	"bufio"
	"encoding/json"

	"github.com/bozso/gomma/settings"
	"github.com/bozso/gotoolbox/path"
)

type ExecutorPayload struct {
	Tag  string `json:"type" toml:"type"`
	Data []byte `json:"data" toml:"data"`
}

type ContextConfig struct {
	Logfile   string          `json:"logfile" toml:"logfile"`
	GammaPath string          `json:"gamma_dir" toml:"gamma_dir"`
	Executor  ExecutorPayload `json:"executor" toml:"executor"`
}

type Context struct {
	Logger        io.Writer
	GammaCommands settings.Commands
	/// TODO: add executor after merge
}

func (c *Context) Set(s string) (err error) {
	f, err := OsFS.Open(s)
	if err != nil {
		return
	}
	defer f.Close()

	var payload ContextPayload
	err = json.NewDecoder(bufio.NewReader(f)).Decode(&payload)
	if err != nil {
		return
	}

	return nil
}
