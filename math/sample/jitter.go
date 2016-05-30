package sample

import (
	"math/rand"
)

// Jitter2D returns k*k jittered samples in [0,1).
// len(samples) = k * k * 2, laid out as { x_00,y_00, x_01,y_01, .. x_0(k-1), y_0(k-1), .. x_(k-1)(k-1), y_(k-1)(k-1) }
func Jitter2D(k int) (samples []float64) {
	d := 1.0 / float64(k)
	u := float64(0)
	for j := 0; j < k; j++ {
		v := float64(0)
		for i := 0; i < k; i++ {
			r0, r1 := rand.Float64(), rand.Float64()
			samples = append(samples, 1.0-(u+d*(1.0-r0)))
			samples = append(samples, 1.0-(v+d*(1.0-r1)))
			v += d
		}
		u += d
	}

	// log.Printf("%v", samples)
	return
}
