package cli

import (
    "bufio"
    "encoding/json"

    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gomma/settings"
)

type ExecutorPayload struct {
    Tag  string `json:"type"`
    Data []byte `json:"data"`
}

type ContextConfig struct {
    Logfile   path.File       `json:"logfile"`
    GammaPath path.Dir        `json:"gamma_dir"`
    Executor  ExecutorPayload `json:"executor"`
}

type Context struct {
    Logger io.Writer
    GammaCommands settings.Commands
    /// TODO: add executor after merge
}

func (c *Context) Set(s string) (err error) {
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
    err = json.NewDecoder(bufio.NewReader(reader)).Decode(&payload)
    if err != nil {
        return
    }



    return nil
}
