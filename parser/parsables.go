package parser

type Parser interface {
	ParseInt(string, Base, BitSize) (int64, error)
	ParseUInt(string, Base, BitSize) (uint64, error)
	ParseFloat(string, BitSize) (float64, error)
}

func Int8(p Parser, s string, b Base) (i int8, err error) {
	v, err := p.ParseInt(s, b, 8)
	if err != nil {
		return
	}

	return int8(v), nil
}
