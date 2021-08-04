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

func MustHave(get Getter, key string) (err error) {
	err = nil
	if !get.HasKey(key) {
		err = &MissingKey{
			Key: key,
		}
	}

	return
}
