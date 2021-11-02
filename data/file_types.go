package data

type TypeEnsurer struct {
	dataTypes []Kind
}

func (t TypeEnsurer) ValidateMeta(m Meta) (err error) {
	return m.MustBeOfType(t.dataTypes...)
}

func NewTypeEnsurer(dtypes ...Kind) (t TypeEnsurer) {
	return TypeEnsurer{
		dataTypes: dtypes,
	}
}

var (
	EnsureComplex = NewTypeEnsurer(KindShortCpx, KindFloatCpx)
	EnsureReal    = NewTypeEnsurer(KindFloat, KindDouble)
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

type Real struct {
	File
}

func (r Real) Validate() (err error) {
	return r.File.Meta.MustBeReal()
}

func LoadReal(l Loader, p PathWithPar) (r Real, err error) {
	f, err := LoadDefault(l, p, EnsureReal)
	if err != nil {
		return
	}

	return Real{f}, nil
}

type Complex struct {
	File
}

func (c Complex) Validate() (err error) {
	return c.File.Meta.MustBeComplex()
}

func LoadComplex(l Loader, p PathWithPar) (c Complex, err error) {
	f, err := LoadDefault(l, p, EnsureComplex)
	if err != nil {
		return
	}

	return Complex{f}, nil
}
