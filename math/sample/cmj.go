package sample

import (
	"math"
	"math/rand"
)

func permute(i, l, p int) int {
	w := l - 1

	w |= w >> 1
	w |= w >> 2
	w |= w >> 4
	w |= w >> 8
	w |= w >> 16

	for {
		i ^= p
		i *= 0xe170893d
		i ^= p >> 16
		i ^= (i & w) >> 4
		i ^= p >> 8
		i *= 0x0929eb3f
		i ^= p >> 23
		i ^= (i & w) >> 1
		i *= 1 | (p >> 27)
		i *= 0x6935fa69
		i ^= (i & w) >> 11
		i *= 0x74dcb303
		i ^= (i & w) >> 2
		i *= 0x9e501cc3
		i ^= (i & w) >> 2
		i *= 0xc860a3df
		i &= w
		i ^= i >> 5

		if i < l {
			break
		}
	}

	return (i + p) % l
}

/*
CMJ1 implements correlated multi-jittering.

Implementation of http://graphics.pixar.com/library/MultiJitteredSampling/paper.pdf
*/
func CMJ1(s, m, n, p int) (x, y float64) {
	sx := float64(permute(s%m, m, p*0xa511e9b3))
	sy := float64(permute(s/m, n, p*0x63d83595))

	jx := rand.Float64()
	jy := rand.Float64()

	x = (float64(s%m) + (sy+jx)/float64(n)) / float64(m)
	y = (float64(s/m) + (sx+jy)/float64(m)) / float64(n)
	return
}

/*
CMJ2 implements correlated multi-jittering.

Implementation of http://graphics.pixar.com/library/MultiJitteredSampling/paper.pdf
*/
func CMJ2(s, N, p int) (x, y float64) {
	a := 1.0
	m := int(math.Sqrt(float64(N) * a))
	n := (N + m - 1) / m

	s = int(permute(s, N, p*0x51633e2d))
	//	sx := float64(permute(s%m, m, p*0xa511e9b3))
	//	sy := float64(permute(s/m, n, p*0x63d83595))
	sx := float64(permute(s%m, m, p*0x68bc21eb))
	sy := float64(permute(s/m, n, p*0x368cc8b7))

	jx := rand.Float64()
	jy := rand.Float64()

	x = (sx + (sy+jx)/float64(n)) / float64(m)
	y = (float64(s) + jy) / float64(N)
	return
}
