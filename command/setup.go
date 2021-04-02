package command

import (
    "encoding/json"

    "github.com/bozso/gotoolbox/meta"
    gmeta "github.com/bozso/gomma/meta"
)

type ExecutorCreator interface {
    CreateExecutor() (Executor, error)
}

type ExecutorCreatorMap map[string]ExecutorCreator

var execs = ExecutorCreatorMap {
    "default": Setup{},
    "debug": DebugConfig{},
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
type ExecutorConfigMap map[string]ExecutorConfig

func (conf ExecutorConfig) ToCreator() (ec ExecutorCreator, err error) {
    ec, ok := execs[conf.Tag]

    if !ok {
        err = gmeta.TagNotFound{
            Tag: conf.Tag,
            Choices: execKeys,
        }
        return
    }

    err = json.Unmarshal(conf.Data, &ec)
    return
}

func (conf ExecutorConfig) CreateExecutor() (e Executor, err error) {
    ec, err := conf.ToCreator()
    if err != nil {
        return
    }

    e, err = ec.CreateExecutor()
    return
}
