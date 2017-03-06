package texture

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

const maxProbes = 16

func sqr(x float32) float32 { return x * x }

/*
Ann= (∂v/∂x)^2+ (∂v/∂y)^2
Bnn= –2 * (∂u/∂x* ∂v/∂x+ ∂u/∂y* ∂v/∂y)
Cnn= (∂u/∂x)^2+ (∂u/∂y)^2
F= Ann*Cnn– Bnn^2/4
A= Ann/F
B= Bnn/F
C= Cnn/F
*/

// SampleFeline returns the filtered RGB value for the given texture file.  Accepts
// normalized texture coordinates in sc.U & sc.V and texture derivatives Dduvdx&dy.
// WRL-99-1
func SampleFeline(filename string, sc *core.ShaderContext) (c [3]float32) {
	textures := texStore.Load().(TexStore)
	img := textures[filename]

	if img == nil {
		img2, err := cacheMiss(filename)

		if err != nil {
			return
		}

		img = img2
	}

	Dduvdx := m.Vec2Scale(sc.Image.PixelDelta[0], sc.Dduvdx)
	Dduvdy := m.Vec2Scale(sc.Image.PixelDelta[1], sc.Dduvdy)

	// NOTE: we need image coordinates for Feline to work.
	// TODO: tidy this all up, the mip mapping and texture structures are
	// a complete mess.

	Dduvdx[0] = Dduvdx[0] * float32(img.w)
	Dduvdy[0] = Dduvdy[0] * float32(img.w)

	Dduvdx[1] = Dduvdx[1] * float32(img.h)
	Dduvdy[1] = Dduvdy[1] * float32(img.h)

	Ann := Dduvdx[1]*Dduvdx[1] + Dduvdy[1]*Dduvdy[1]
	Bnn := -2 * (Dduvdx[0]*Dduvdx[1] + Dduvdy[0]*Dduvdy[1])
	Cnn := Dduvdx[0]*Dduvdx[0] + Dduvdy[0]*Dduvdy[0]
	F := Ann*Cnn - (Bnn * Bnn / 4)

	A := Ann / F
	B := Bnn / F
	C := Cnn / F

	// TODO: lots of this can be approximated as we don't need high accuracy due to
	// the heavy weighted filter.

	root := m.Sqrt(sqr(A-C) + sqr(B))
	Aprm := (A + C - root) / 2
	Cprm := (A + C + root) / 2

	majorRadius := m.Sqrt(1 / Aprm)
	minorRadius := m.Sqrt(1 / Cprm)

	theta := m.Atan(B/(A-C)) / 2

	// If theta is angle of minor axis make it andlge of major
	if A > C {
		theta = theta + m.Pi/2
	}

	// Clamp to 1 pixel
	minorRadius = m.Max(minorRadius, 1)
	majorRadius = m.Max(majorRadius, 1)

	// desired number of probes
	fProbes := 2*(majorRadius/minorRadius) - 1
	iProbes := m.Floor(fProbes + 0.5)

	iProbes = m.Min(iProbes, maxProbes)

	if iProbes < fProbes {
		minorRadius = 2 * majorRadius / (iProbes + 1)
	}

	levelOfDetail := m.Log2(minorRadius)

	if levelOfDetail > float32(img.mipmap.MaxLevelOfDetail()) {
		levelOfDetail = float32(img.mipmap.MaxLevelOfDetail())
		iProbes = 1
	}

	if levelOfDetail < 0 {
		levelOfDetail = 0
	}

	lineLength := 2 * (majorRadius - minorRadius)

	dU := m.Cos(theta) * lineLength / (iProbes - 1)
	dV := m.Sin(theta) * lineLength / (iProbes - 1)

	nProbes := int(iProbes)

	if nProbes == 1 {
		// Avoid NaNs
		dU = 0
		dV = 0
	}

	n := float32(-(nProbes - 1))

	alpha := float32(0.6)

	var accum [3]float32
	var accumWeight float32

	for i := 0; i < nProbes; i++ {
		u := float32(img.w)*sc.U + (n/2)*dU
		v := float32(img.h)*sc.V + (n/2)*dV

		//d := float32(n) / 2 * m.Sqrt(sqr(dU)+sqr(dV)) / majorRadius
		d2 := (sqr(n) / 4) * (sqr(dU) + sqr(dV)) / sqr(majorRadius)
		relativeWeight := m.Exp(-alpha * d2)

		sample := img.mipmap.TrilinearSample(u/float32(img.w), v/float32(img.h), levelOfDetail)

		for k := range accum {
			accum[k] += (sample[k] / 255.0) * relativeWeight

		}

		accumWeight += relativeWeight

		n += 2
	}

	for k := range accum {
		c[k] = accum[k] / accumWeight
	}

	//c[0] = levelOfDetail * 10 //m.Floor(levelOfDetail) / float32(img.mipmap.MaxLevelOfDetail())

	return

}
