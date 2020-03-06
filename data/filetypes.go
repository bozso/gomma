package data

type FloatFile struct {
    File
}

func (f FloatFile) Validate() (err error) {
    return f.TypeCheck(Float, Double)
}

type ComplexFile struct {
    File
}

func (c ComplexFile) Validate() (err error) {
    return c.TypeCheck(ShortCpx, FloatCpx)
}
