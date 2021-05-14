package command

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/bozso/gomma/meta"
	"github.com/bozso/gomma/stream"
)

const payload = `
{
    "debug": {
        "debug": {
            "logfile": "/tmp/test.log"
        }
    },
    "default": { "default": {} }
}
`

type (
	Payloads           map[string]meta.Payload
	ExecutorCreatorMap map[string]ExecutorCreator
)

var reference = ExecutorCreatorMap{
	"default": Setup{},
	"debug": DebugConfig{
		Logfile: stream.Config{
			Mode:    stream.Path,
			Logfile: "/tmp/test.log",
		},
	},
}

func DecodeConfigs(t *testing.T) (confs Payloads) {
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
		creator, err := ToCreator(val)
		if err != nil {
			t.Fatalf("could not create creator: %s", err)
		}

		creators[key] = creator
	}

	if !reflect.DeepEqual(reference, creators) {
		t.Errorf("expected %v and %v to be equal", confs, reference)
	}
}
