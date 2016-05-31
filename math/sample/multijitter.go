package sample

import (
	"math"
	"math/rand"
)

// MultiJitter returns k multijittered 2D samples in [0,1).  len(samples) = k*2, laid out
// as { x_0,y_0,x_1,y_1,x_2,y_2,... }
//
// Deprecated: Not used anywhere and QMC will completely redo sampling strategies.
func MultiJitter(k int) (samples []float64) {
	samples = make([]float64, 0, k*2)
	n := int(math.Sqrt(float64(k)))
	m := k / n

	// Generate canonical arrangement
	for j := 0; j < n; j++ {
		for i := 0; i < m; i++ {
			x := (float64(i) + (float64(j)+rand.Float64())/float64(n)) / float64(m)
			y := (float64(j) + (float64(i)+rand.Float64())/float64(m)) / float64(n)
			samples = append(samples, x, y)

		}
	}

	// Shuffle x
	for j := 0; j < n; j++ {
		for i := 0; i < m; i++ {
			k := int((float64(j) + rand.Float64()) * float64(n-j))
			samples[(j*m+i)+0], samples[(k*m+i)+0] = samples[(k*m+i)+0], samples[(j*m+i)+0]

		}
	}

	// Shuffle y
	for j := 0; j < n; j++ {
		for i := 0; i < m; i++ {
			k := int((float64(i) + rand.Float64()) * float64(m-i))
			samples[(j*m+i)+1], samples[(k*m+i)+1] = samples[(k*m+i)+1], samples[(j*m+i)+1]

		}
	}
	/*
		for i := 0; i < k; i++ {
			samples = append(samples, rand.Float64(), rand.Float64())
		}*/
	return
}
