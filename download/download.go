package download

import ()

type Query struct {
}

type SceneMeta struct {
}

type Downloader interface {
	QueryImages(Query) ([]SceneMeta, error)
	DownloadImages([]SceneMeta, error)
	DownloadQuery(Query)
}
