package sample

import (
	"math/rand"
	"testing"
	"vermeer/math"
)

func TestSample(t *testing.T) {

	for k := 0; k < 10000; k++ {
		t.Logf("%v", math.Vec3Length(CosineHemisphere(rand.Float64(), rand.Float64())))
	}
}
