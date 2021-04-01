package command

import (
    "encoding/json"

    "github.com/bozso/gotoolbox/meta"
    "github.com/bozso/gotoolbox/enum"
)

type ExecutorConfig interface {
    SetupExecutor() (Executor, error)
}

type ExecutorMap map[string]Executor

var execs = ExecutorMap {
    "default": NewExecute(),
}

var (
    setupKeys []string
    s meta.Startup
    _ = s.Do(&meta.MapKeysGet{
        Map: execs,
        Keys: setupKeys,
    })
)

type TagNotFound struct {
    Tag string
    Choices enum.StringSet
}

func selectSetup(execs ExecutorMap, b []byte) (e Executor, err error) {
    var payload struct {
        Tag string  `json:"type"`
        Data []byte `json:"data"`
    }

    if err = json.Unmarshal(b, &payload); err != nil {
        return
    }

    e, ok := execs[payload.Tag]

    if !ok {
        // TODO: set error message
        return
    }


    return
}

/*
func (e *ExecutorFromJSON) UnmarshalJSON(b []byte) (err error) {
    setup, err := selectSetup(setups, b)
    if err != nil {
        return
    }
    e.ExecutorSetup = setup
    return
}
*/
