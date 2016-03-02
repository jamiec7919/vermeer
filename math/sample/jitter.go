package sample

import (
	"math/rand"
)

// k * k samples * 2
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
