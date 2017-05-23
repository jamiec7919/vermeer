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
		0.41231515, 0.3575761, 0.1804375,
		0.2126729, 0.7151522, 0.0721750,
		0.019333892, 0.11919198, 0.95063042,
	},
	/*
		// Bradford adapted to D50
		mInv: [9]float32{
			0.4360747, 0.3850649, 0.1430804,
			0.2225045, 0.7168786, 0.0606169,
			0.0139322, 0.0971045, 0.7141733,
		},*/
}

var SRGB = sRGB
