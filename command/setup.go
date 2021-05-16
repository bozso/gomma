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
	pl, err := json.UnmarshalPayloads(b)
	if err != nil {
		return
	}

	var ex Executor = nil
	pl.ForEach(func(key string, val json.RawMsg) (err error) {
		err = nil
		switch key {
		case "default":
			ex = NewExecute()
		case "debug":
			var debug Debug
			err = stdJson.Unmarshal(val, &debug)
			ex = debug
		}
		return err
	})

	if ex == nil {
		err = pl.NoMatchingTag(tags)
	} else {
		e.Executor = ex
		err = nil
	}

	return
}
