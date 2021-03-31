package command

import (
    "encoding/json"

    "github.com/bozso/gotoolbox/meta"
)

type ExecutorConfig interface {
    SetupExecutor() (Executor, error)
}

type ExecutorConfigs map[string]ExecutorConfig

var setups = ExecutorConfigs {
    "default": DefaultExecutor{},
}

var (
    setupKeys []string
    s meta.Startup
    _ = s.Do(&meta.MapKeysGet{
        Map: setups,
        Keys: setupKeys,
    })
)

type DefaultExecutor struct {}

func (_ DefaultExecutor) SetupExecutor() (ex Executor, err error) {
    return NewExecute(), nil
}

type DebugExecutor struct {


}


func selectSetup(confs ExecutorConfigs, b []byte) (es ExecutorConfig, err error) {
    var payload struct {
        tag string  `json:"type"`
        data []byte `json:"data"`
    }

    if err = json.Unmarshal(b, payload); err != nil {
        return
    }

    if dataType, ok := confs[payload.tag]; !ok {
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
