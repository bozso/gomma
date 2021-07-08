package data

import (
	"time"

	"git.sr.ht/~istvan_bozso/shutil/path"
)

type Typer interface {
	Type() Type
}

type Like interface {
	Path() path.Path
	Meta() Meta
}
