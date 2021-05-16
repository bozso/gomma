package command

import (
	stdJson "encoding/json"
	//"reflect"
	"strings"
	"testing"

	json "git.sr.ht/~istvan_bozso/sert/json"

	"github.com/bozso/gomma/stream"
)

var payload = `
{
    "debug": {
        "debug": {
            "logfile": "stdout",
            "formatter": "default"
        }
    },
    "default": { "default": {} }
}
`

var reference = map[string]Executor{
	"default": NewExecute(),
	"debug": Debug{
		Out: stream.Stdout(),
		Fmt: FormatterJSON{LineFormat},
	},
}

func DecodeConfigs(t *testing.T) (confs json.Payloads) {
	dec := stdJson.NewDecoder(strings.NewReader(payload))

	if err := dec.Decode(&confs); err != nil {
		t.Fatalf("could not decode %s into a config map: %s", payload, err)
	}
	return
}

/*
func TestDecodeSetup(t *testing.T) {
	confs := DecodeConfigs(t)

	for key, val := range confs {
		var ex ExecutorJSON
		err := ex.UnmarshalJSON(val)
		if err != nil {
			t.Fatalf("unmarshaling failed: %s", err)
		}
		now, ref := ex.Executor, reference[key]

		if !reflect.DeepEqual(now, ref) {
			t.Errorf("expected value \n%#v\n to be equal to \n%#v\n", now, ref)
		}
	}
}
*/
