package command

import (
    "testing"
    "strings"
    "reflect"
    "encoding/json"

    "github.com/bozso/gomma/stream"
)

const payload = `
{
    "debug": {
        "type": "debug",
        "data": {
            "logfile": "/tmp/test.log"
        }
    },
    "default": {
        "type": "default"
    }
}
`

var reference = ExecutorCreatorMap {
    "default": Setup{},
    "debug": DebugConfig {
        Logfile: stream.Config {
            Mode: stream.Path,
            Data: "/tmp/test.log",
        },
    },
}

func DecodeConfigs(t *testing.T) (confs ExecutorConfigMap) {
    dec := json.NewDecoder(strings.NewReader(payload))

    if err := dec.Decode(&confs); err != nil {
        t.Fatalf("could not decode %s into a config map: %s", payload, err)
    }
    return
}

func TestDecodeSetup(t *testing.T) {
    confs := DecodeConfigs(t)

    creators := make(ExecutorCreatorMap)

    for key, val := range confs {
        creator, err := val.ToCreator()
        if err != nil {
            t.Fatalf("could not create creator: %s", err)
        }

        creators[key] = creator
    }


    if !reflect.DeepEqual(reference, creators) {
        t.Errorf("expected %v and %v to be equal", confs, reference)
    }
}
