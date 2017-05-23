package colour

import (
	"math/rand"
	"testing"
)

const Count = 100000

func TestSpectrum(t *testing.T) {

	rgb := RGB{1, 1, 1}
	rgb2 := RGB{1, 0.1, 0.1}

	xyz := sRGB.RGBToXYZ(rgb)
	xyz = ChromaticAdjust(XYZScalingD65ToE, xyz)
	//xyz = [3]float32{0.8, 0.8, 0.8}
	t.Logf("XYZ: %v RGB: %v", xyz, sRGB.XYZToRGB(xyz[0], xyz[1], xyz[2]))

	for i := 0; i < 50; i++ {
		lambda := float32(((LambdaMax-LambdaMin)/50)*i) + LambdaMin

		v := spectrumXYZToP(lambda, xyz)

		t.Logf("%v", v)
	}
	//xyz2 := sRGB.RGBToXYZ(rgb2)

	var c RGB
	var base RGB

	for i := 0; i < Count; i++ {
		tt := Spectrum{Lambda: (LambdaMax-LambdaMin)*rand.Float32() + LambdaMin}
		tt2 := Spectrum{Lambda: tt.Lambda}
		tt.FromRGB(rgb)
		tt2.FromRGB(rgb2)
		base.Add(tt.ToXYZ())

		//t.Logf("1: %v ", tt)
		tt.Mul(tt2)

		c.Add(tt.ToXYZ())
		//t.Logf("%v %v %v ", xyz, tt, tt.ToXYZ())
		//t.Logf("2: %v ", tt2)
		//t.Logf("1*2: %v ", tt.ToRGB())
	}

	c.Scale(1.0 / Count)
	base.Scale(1.0 / Count)

	base = ChromaticAdjust(XYZScalingEToD65, base)
	c = ChromaticAdjust(XYZScalingEToD65, c)

	t.Logf("%v %v %v", c, sRGB.XYZToRGB(c[0], c[1], c[2]), sRGB.XYZToRGB(base[0], base[1], base[2]))
}
