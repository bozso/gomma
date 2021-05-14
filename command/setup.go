package command

import (
	stdJson "encoding/json"
	json "git.sr.ht/~istvan_bozso/sert/json"
)

type ExecutorJSON struct {
	Executor
}

var tags = []string{
	"default", "debug",
}

func (e *ExecutorJSON) UnmarshalJSON(b []byte) (err error) {
	var pl json.Payloads
	if err = stdJson.Unmarshal(b, &pl); err != nil {
		return
	}

	var ex Executor = nil
	for key, val := range pl {
		switch key {
		case "default":
			ex = NewExecute()
		case "debug":
			var debug Debug
			err = stdJson.Unmarshal(val, &debug)
			ex = debug
		}
		if err != nil {
			return
		}
	}
	if ex == nil {
		err = pl.NoMatchingTag(tags)
	} else {
		e.Executor = ex
		err = nil
	}

	return
}
