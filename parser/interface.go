package parser

type Getter interface {
	GetParsed(key string) (string, error)
}

type Setter interface {
	SetParsed(key, value string) error
}

type MutGetter interface {
	Setter
	Getter
}
