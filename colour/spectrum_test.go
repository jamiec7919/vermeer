package colour

import (
	"math/rand"
	"testing"
)

const Count = 100000

func TestSpectrum(t *testing.T) {

	rgb := RGB{1, 1, 1}
	rgb2 := RGB{1, 0.1, 0.1}
	//xyz := sRGB.RGBToXYZ(rgb)
	//xyz2 := sRGB.RGBToXYZ(rgb2)

	var c RGB

	for i := 0; i < Count; i++ {
		tt := Spectrum{Lambda: (LambdaMax-LambdaMin)*rand.Float32() + LambdaMin}
		tt2 := Spectrum{Lambda: tt.Lambda}
		tt.FromRGB(rgb)
		tt2.FromRGB(rgb2)

		//t.Logf("1: %v ", tt)
		tt.Mul(tt2)

		c.Add(tt.ToXYZ())
		//t.Logf("%v %v %v ", xyz, tt, tt.ToXYZ())
		//t.Logf("2: %v ", tt2)
		//t.Logf("1*2: %v ", tt.ToRGB())
	}

	c.Scale(1.0 / Count)
	t.Logf("%v %v", c, sRGB.XYZToRGB(c[0], c[1], c[2]))
}
