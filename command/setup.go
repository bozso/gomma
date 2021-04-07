package command

import (
    "encoding/json"

    gmeta "github.com/bozso/gomma/meta"
)

type ExecutorCreator interface {
    CreateExecutor() (Executor, error)
}

type ExecutorConfig gmeta.Config

var execKeys = []string{"default", "debug"}

func (conf ExecutorConfig) ToCreator() (ec ExecutorCreator, err error) {
    switch conf.Tag {
    case "default":
        var setup Setup
        err = json.Unmarshal(conf.Data, &setup)
        ec = setup
    case "debug":
        var dc DebugConfig
        err = json.Unmarshal(conf.Data, &dc)
        ec = dc
    default:
        err = gmeta.TagNotFound {
            Tag: conf.Tag,
            Choices: execKeys,
        }
        return
    }

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
