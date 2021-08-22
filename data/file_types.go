package data

type TypeEnsurer struct {
	dataTypes []Type
}

func (t TypeEnsurer) ValidateMeta(m Meta) (err error) {
	return m.MustBeOfType(t.dataTypes...)
}

func NewTypeEnsurer(dtypes ...Type) (t TypeEnsurer) {
	return TypeEnsurer{
		dataTypes: dtypes,
	}
}

var (
	EnsureComplex = NewTypeEnsurer(ShortCpx, FloatCpx)
	EnsureReal    = NewTypeEnsurer(Float, Double)
)

func LoadAndValidate(l Loader, p PathWithPar, pk ParamKeys, v MetaValidator) (f File, err error) {
	mp := ParseAndValidate{
		parser:    pk,
		validator: v,
	}

	return l.LoadFile(p, mp)
}

func LoadDefault(l Loader, p PathWithPar, v MetaValidator) (f File, err error) {
	return LoadAndValidate(l, p, DefaultKeys, v)
}

func LoadComplex(l Loader, p PathWithPar) (f File, err error) {
	return LoadDefault(l, p, EnsureComplex)
}

func LoadReal(l Loader, p PathWithPar) (f File, err error) {
	return LoadDefault(l, p, EnsureReal)
}
