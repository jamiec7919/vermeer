// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

// Implements Smits99 RGB->Spectrum conversion.

const (
	smitsLambdaMin float32 = 380
	smitsLambdaMax         = 720
)

type smitsSpectrum [10]float32

func (s *smitsSpectrum) eval(lambda float32) float32 {
	if lambda < smitsLambdaMin || lambda >= smitsLambdaMax {
		return 0
	}

	bin := int(((lambda - smitsLambdaMin) / (smitsLambdaMax - smitsLambdaMin)) * 10.0)

	return s[bin]

}

/*
var (
	Rxy = ColXY{.64, .33}
	Gxy = ColXY{.3, .6}
	Bxy = ColXY{.15, .06}
	Wxy = ColXY{.333, .333}
	WY  = float32(106.8)
)
*/

var (
	white   = smitsSpectrum{1, 1, 0.999, 0.9993, 0.9992, 0.9998, 1, 1, 1, 1}
	cyan    = smitsSpectrum{0.9710, 0.9426, 1.0007, 1.0007, 1.0007, 0.1564, 0, 0, 0, 0}
	magenta = smitsSpectrum{1, 1, 0.9685, 0.2229, 0, 0.0458, 0.8369, 1, 1, 0.9959}
	yellow  = smitsSpectrum{0.0001, 0, 0.1088, 0.6651, 1, 1, 0.9996, 0.9586, 0.9685, 0.9840}
	red     = smitsSpectrum{0.1012, 0.0515, 0, 0, 0, 0, 0.8325, 1.0149, 1.0149, 1.0149}
	green   = smitsSpectrum{0, 0, 0.0273, 0.7937, 1, 0.9418, 0.1719, 0, 0, 0.0025}
	blue    = smitsSpectrum{1, 1, 0.8916, 0.3323, 0, 0, 0.0003, 0.0369, 0.0483, 0.0496}
)

// RGBToSpectrumSmits99 converts an RGB triple to a spectrum at a given wavelength
func RGBToSpectrumSmits99(r, g, b, lambda float32) (c float32) {
	if r <= g && r <= b {

		c += r * white.eval(lambda)

		if g <= b {
			c += (g - r) * cyan.eval(lambda)
			c += (b - g) * blue.eval(lambda)
		} else {
			c += (b - r) * cyan.eval(lambda)
			c += (g - b) * green.eval(lambda)
		}
	} else if g <= r && g <= b {
		c += g * white.eval(lambda)

		if r <= b { // FIXME I'm not sure these should be yellow/magenta
			c += (r - g) * magenta.eval(lambda)
			c += (b - r) * blue.eval(lambda)
		} else {
			c += (b - g) * magenta.eval(lambda)
			c += (r - b) * red.eval(lambda)
		}

	} else { // b <= red && b <= g
		c += b * white.eval(lambda)

		if r <= g { // FIXME I'm not sure these should be yellow/magenta
			c += (r - b) * yellow.eval(lambda)
			c += (g - r) * green.eval(lambda)
		} else {
			c += (g - b) * yellow.eval(lambda)
			c += (r - g) * red.eval(lambda)
		}
	}

	return
}
