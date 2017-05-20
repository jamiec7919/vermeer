// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

// sRGB is a concrete ColourSpace for the sRGB space.
var sRGB = Space{
	m: [9]float32{
		3.2404542, -1.5371385, -0.4985314,
		-0.9692660, 1.8760108, 0.0415560,
		0.0556434, -0.2040259, 1.0572252,
	},
	mInv: [9]float32{
		0.41231515, 0.3576, 0.1805,
		0.2126, 0.7152, 0.0722,
		0.01932727, 0.1192, 0.95063333,
	},
}

var SRGB = sRGB
