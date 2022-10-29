package lib

import (
	"math/rand"

	"github.com/XANi/loremipsum"
)

func RandInt(l, r int) int {
	if l > r {
		return l
	}
	return l + rand.Int()%(r-l+1)
}

func Sample[T any](a []T) T {
	return a[RandInt(0, len(a)-1)]
}

func Lorem() *loremipsum.LoremIpsum {
	return loremipsum.NewWithSeed(rand.Int63())
}
