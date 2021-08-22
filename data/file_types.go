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
