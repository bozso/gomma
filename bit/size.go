package bit

import (
	"fmt"
)

type Size int

type empty struct{}

type Sizes map[int]empty

func NewSizes(values ...int) (s Sizes) {
	s = make(Sizes)
	for _, value := range values {
		s[value] = empty{}
	}

	return
}

func panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

func (ss Sizes) Get(bitSize int) (s Size) {
	_, ok := ss[bitSize]

	if !ok {
		panicf("invalid bitsize '%d'", bitSize)
	}

	return Size(bitSize)
}

var intSizes = NewSizes(8, 16, 32, 64)

func IntSize(bitSize int) (s Size) {
	return intSizes.Get(bitSize)
}
