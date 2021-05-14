package command

import (
	"errors"

	"git.sr.ht/~istvan_bozso/sert/json"
	"github.com/bozso/gomma/meta"
)

type ExecutorCreator interface {
	CreateExecutor() (Executor, error)
}

var tags = [...]string{"default", "debug"}

func ToCreator(pl json.Payload) (ec ExecutorCreator, err error) {
	for _, tag := range tags {
		switch tag {
		case "default":
			var setup Setup
			err = pl.Unmarshal(tag, &setup)
			ec = setup
		case "debug":
			var dc DebugConfig
			err = pl.Unmarshal(tag, &dc)
			ec = dc
		}
		if errors.Is(err, meta.MissingTagError) {
			continue
		} else {
			break
		}
	}

	return
}

func CreateExecutor(pl json.Payload) (e Executor, err error) {
	ec, err := ToCreator(pl)
	if err != nil {
		return
	}

	e, err = ec.CreateExecutor()
	return
}
