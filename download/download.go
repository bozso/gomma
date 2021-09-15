package download

import (
	"time"

	"github.com/bozso/gomma/geometry"
)

type Query struct {
	AOI geometry.AreaOfInterest
}

type SceneMeta struct {
	Date time.Time
	URL  string
}

type Downloader interface {
	QueryImages(Query) ([]SceneMeta, error)
	DownloadImages([]SceneMeta, error)
	DownloadQuery(Query)
}
