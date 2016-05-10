// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

/*

  Uses the idea of Hero Wavelength. (add ref..)

  NOTES: pure colours don't work very well, make sure to always mix at least some of each primary otherwise
  terrible speckling will occur.  You have been warned.  Should add a verification function that will
  make input RGB values 'reasonable'
*/

const (
	LAMBDA_MIN = 450
	LAMBDA_MAX = 750
	LAMBDA_N   = 4
)

const lambda_bar = LAMBDA_MAX - LAMBDA_MIN

type Spectrum struct {
	C      [LAMBDA_N]float32
	Lambda float32
}

func (wv *Spectrum) SetZero() {
	for k := 0; k < 4; k++ {
		wv.C[k] = 0
	}

}

func (wv *Spectrum) Set(v float32) {
	for k := 0; k < 4; k++ {
		wv.C[k] = v
	}

}

func (wv *Spectrum) FromRGB(r, g, b float32) error {
	for k := 0; k < 4; k++ {
		wv.C[k] = RGBToSpectrumSmits99(r, g, b, wv.Wavelength(k))
	}
	return nil
}

func (wv *Spectrum) ToRGB() (r, g, b float32) {
	var x, y, z float32

	// Treat s as a line spectrum, therefore integrals required are: Glassner 1987
	// (How to Derive a Spectrum From an RGB Triple)

	for i := 0; i < LAMBDA_N; i++ {
		x += wv.C[i] * Cie_x(wv.Wavelength(i))
		y += wv.C[i] * Cie_y(wv.Wavelength(i))
		z += wv.C[i] * Cie_z(wv.Wavelength(i))
	}

	// Now transform XYZ to RGB.  Should make this selectable as to which RGB.
	r, g, b = ColourSpace_sRGB.XYZToRGB(x, y, z)
	return
}

func (wv *Spectrum) Mul(other Spectrum) {
	// if wv.Lambda != other.Lambda IS ERROR
	wv.C[0] *= other.C[0]
	wv.C[1] *= other.C[1]
	wv.C[2] *= other.C[2]
	wv.C[3] *= other.C[3]
}

func (wv *Spectrum) Scale(s float32) {
	// if wv.Lambda != other.Lambda IS ERROR
	wv.C[0] *= s
	wv.C[1] *= s
	wv.C[2] *= s
	wv.C[3] *= s
}

func (wv *Spectrum) Add(other Spectrum) {
	// if wv.Lambda != other.Lambda IS ERROR
	wv.C[0] += other.C[0]
	wv.C[1] += other.C[1]
	wv.C[2] += other.C[2]
	wv.C[3] += other.C[3]
}

// j = 0..LAMBDA_N
func (wv *Spectrum) Wavelength(j int) (v float32) {

	v = (wv.Lambda - LAMBDA_MIN + (float32(j)/LAMBDA_N)*lambda_bar)

	if v >= lambda_bar {
		v -= lambda_bar
	}

	v += LAMBDA_MIN

	return
}
