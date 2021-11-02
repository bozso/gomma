package geometry

type Index int

type Indices struct {
	indices []Index
}

func NewIndices(ints ...int) (ind Indices) {
	n := len(ints)
	in := make([]Index, n)

	for ii, int_ := range ints {
		in[ii] = Index(int_)
	}

	return Indices{
		indices: in,
	}
}

func (in Indices) At(ii int) Index {
	return in.indices[ii]
}

func (in Indices) Has(idx Index) (b bool) {
	b = false
	for _, curr := range in.indices {
		if curr == idx {
			b = true
			break
		}
	}
	return
}

func (in Indices) HasAll(indices []Index) (b bool) {
	b = true
	for _, idx := range indices {
		if !in.Has(idx) {
			b = false
			break
		}
	}
	return
}

func (in Indices) HasAny(indices []Index) (b bool) {
	b = false
	for _, idx := range indices {
		if in.Has(idx) {
			b = true
			break
		}
	}
	return
}
