package data

func (f File) EnsureFloat() (err error) {
    return f.TypeCheck(Double, Float)    
}

func (f File) EnsureComplex() (err error) {
    return f.TypeCheck(ShortCpx, FloatCpx)    
}
