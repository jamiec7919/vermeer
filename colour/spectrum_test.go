package colour

import (
	"math/rand"
	"testing"
)

const Count = 1000

func TestSpectrum(t *testing.T) {

	rgb := RGB{1, 1, 1}
	xyz := sRGB.RGBToXYZ(rgb)

	var c RGB

	for i := 0; i < Count; i++ {
		tt := Spectrum{Lambda: (LambdaMax-LambdaMin)*rand.Float32() + LambdaMin}
		tt.FromRGB(rgb)

		c.Add(tt.ToXYZ())
		//t.Logf("%v %v %v ", xyz, tt, tt.ToXYZ())
		t.Logf("%v %v %v %v", tt.Wavelength(0), tt.Wavelength(1), tt.Wavelength(2), tt.Wavelength(3))
	}

	c.Scale(1.0 / Count)
	t.Logf("%v %v %v", c, xyz, sRGB.XYZToRGB(xyz[0], xyz[1], xyz[2]))
}
