package command

import (
    "encoding/json"
)

type ExecutorSetup interface {
    SetupExecutor() (Executor, error)
}

type ExecutorSetups map[string]ExecutorSetup

var setups = ExecutorSetups{
    "default": DefaultExecutorSetup{},
}

var setupKeys = 

type DefaultExecutorSetup struct {}

func (_ DefaultExecutorSetup) SetupExecutor() (ex Executor, err error) {
    return NewExecute(), nil
}

type DebugExecutorSetup struct {

}


func selectSetup(setups ExecutorSetups, b []byte) (es ExecutorSetup, err error) {
    var payload struct {
        tag string  `json:"type"`
        data []byte `json:"data"`
    }

    if err = json.Unmarshal(b, payload); err != nil {
        return
    }

    if dataType, ok := setups[payload.tag]; !ok {
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
