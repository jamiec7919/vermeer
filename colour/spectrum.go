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

// Constants for min and maximum wavelength and number of wavelength samples in use for the
// hero-wavelength.
const (
	LAMBDA_MIN = 450
	LAMBDA_MAX = 750
	LAMBDA_N   = 4
)

const lambda_bar = LAMBDA_MAX - LAMBDA_MIN

// Spectrum represents a line spectrum using the hero-wavelength idea.
type Spectrum struct {
	C      [LAMBDA_N]float32
	Lambda float32
}

// SetZero sets the spectrum to 0.
func (wv *Spectrum) SetZero() {
	for k := 0; k < 4; k++ {
		wv.C[k] = 0
	}

}

// Set sets the spectrum to a constant value.
func (wv *Spectrum) Set(v float32) {
	for k := 0; k < 4; k++ {
		wv.C[k] = v
	}

}

// FromRGB ses the spectrum from the RGB values.
func (wv *Spectrum) FromRGB(r, g, b float32) error {
	for k := 0; k < 4; k++ {
		wv.C[k] = RGBToSpectrumSmits99(r, g, b, wv.Wavelength(k))
	}
	return nil
}

// ToRGB converts the spectrum to RGB.
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

// Mul multiplies the spectrum by other.  Both spectra must represent the same wavelength.
func (wv *Spectrum) Mul(other Spectrum) {
	// if wv.Lambda != other.Lambda IS ERROR
	wv.C[0] *= other.C[0]
	wv.C[1] *= other.C[1]
	wv.C[2] *= other.C[2]
	wv.C[3] *= other.C[3]
}

// Scale scales the spectrum by s.
func (wv *Spectrum) Scale(s float32) {
	wv.C[0] *= s
	wv.C[1] *= s
	wv.C[2] *= s
	wv.C[3] *= s
}

// Add adds other to the spectrum.  Both spectra must represent the same wavelength.
func (wv *Spectrum) Add(other Spectrum) {
	// if wv.Lambda != other.Lambda IS ERROR
	wv.C[0] += other.C[0]
	wv.C[1] += other.C[1]
	wv.C[2] += other.C[2]
	wv.C[3] += other.C[3]
}

// Wavelength returns the wavelength for index j (see hero-wavelength paper).
// j = 0..LAMBDA_N
func (wv *Spectrum) Wavelength(j int) (v float32) {

	v = (wv.Lambda - LAMBDA_MIN + (float32(j)/LAMBDA_N)*lambda_bar)

	if v >= lambda_bar {
		v -= lambda_bar
	}

	v += LAMBDA_MIN

	return
}
