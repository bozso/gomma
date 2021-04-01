package command

import (
    "encoding/json"

    "github.com/bozso/gotoolbox/meta"
    gmeta "github.com/bozso/gomma/meta"
)

type ExecutorConfig interface {
    ToExecutor() (Executor, error)
}

type ExecutorConfigMap map[string]ExecutorConfig

var execs = ExecutorConfigMap {
    "default": NewExecute(),
}

var (
    execKeys []string
    s meta.Startup
    _ = s.Do(&meta.MapKeysGet{
        Map: execs,
        Keys: execKeys,
    })
)

type ExecutorConfig gmeta.Config

func (ec ExecutorConfig) ToExecutor() (e Executor, err error) {
    e, ok := execs[ec.Tag]

    if !ok {
        err = gmeta.TagNotFound{
            Tag: ec.Tag,
            Choices: execKeys,
        }
        return
    }

    err = json.Unmarshal(ec.Data, &e)
    return
}
