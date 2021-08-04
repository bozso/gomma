package parser

type Getter interface {
	Keys() (keys []string)
	HasKey(key string) (hasKey bool)
	GetParsed(key string) (value string, err error)
}

type Setter interface {
	SetParsed(key, value string) error
}

type MutGetter interface {
	Setter
	Getter
}
