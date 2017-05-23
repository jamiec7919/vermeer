// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package colour provides spectral, hero-wavelength and RGB support functions and types.
package colour

import (
	"github.com/jamiec7919/vermeer/math"
	"log"
)

// Space represents a transform from XYZ to a different space.
type Space struct {
	m    [9]float32 // Matrix to convert XYZ->RGB
	mInv [9]float32 // Matrix to convert RGB->XYZ
}

// XYZToRGB returns the transformed XYZ to RGB
func (s *Space) XYZToRGB(x, y, z float32) (rgb RGB) {
	rgb[0] = x*s.m[0] + y*s.m[1] + z*s.m[2]
	rgb[1] = x*s.m[3] + y*s.m[4] + z*s.m[5]
	rgb[2] = x*s.m[6] + y*s.m[7] + z*s.m[8]

	return
}

// RGBToXYZ returns the transformed RGB to XYZ
func (s *Space) RGBToXYZ(rgb RGB) (xyz [3]float32) {
	xyz[0] = rgb[0]*s.mInv[0] + rgb[1]*s.mInv[1] + rgb[2]*s.mInv[2]
	xyz[1] = rgb[0]*s.mInv[3] + rgb[1]*s.mInv[4] + rgb[2]*s.mInv[5]
	xyz[2] = rgb[0]*s.mInv[6] + rgb[1]*s.mInv[7] + rgb[2]*s.mInv[8]

	return
}

// BuildRGBToXYZMatrix computes the 3 × 3 matrix for converting RGB to XYZ.
// Chromaticity coordinates of an RGB system (xr, yr), (xg, yg) and (xb, yb) and its reference white (XW, YW, ZW).
func BuildRGBToXYZMatrix(xr, yr, xg, yg, xb, yb, Xw, Yw, Zw float32) [9]float32 {
	m := [9]float32{
		xr / yr, xg / yg, xb / yb,
		1, 1, 1,
		(1 - xr - yr) / yr, (1 - xg - yg) / yg, (1 - xb - yb) / yb,
	}

	mInv, _ := math.Matrix3Inverse(m)

	Sr := Xw*mInv[0] + Yw*mInv[1] + Zw*mInv[2]
	Sg := Xw*mInv[3] + Yw*mInv[4] + Zw*mInv[5]
	Sb := Xw*mInv[6] + Yw*mInv[7] + Zw*mInv[8]

	return [9]float32{
		Sr * m[0], Sg * m[1], Sb * m[2],
		Sr * m[3], Sg * m[4], Sb * m[5],
		Sr * m[6], Sg * m[7], Sb * m[8],
	}
}

//BradfordD50ToD65 is the matrix for conversion from illuminant D50 to D65 (Bradford algorithm).
var BradfordD50ToD65 = [...]float32{
	0.9555766, -0.0230393, 0.0631636,
	-0.0282895, 1.0099416, 0.0210077,
	0.0122982, -0.0204830, 1.3299098,
}

//BradfordD65ToD50 is the matrix for conversion from illuminant D65 to D50 (Bradford algorithm).
var BradfordD65ToD50 = [...]float32{
	1.0478112, 0.0228866, -0.0501270,
	0.0295424, 0.9904844, -0.0170491,
	-0.0092345, 0.0150436, 0.752131,
}

// XYZScalingD65ToE is the matrix for conversion from illuminant D65 to E.
var XYZScalingD65ToE = [...]float32{
	1.0521111, 0.0000000, 0.0000000,
	0.0000000, 1.0000000, 0.0000000,
	0.0000000, 0.0000000, 0.9184170,
}

// XYZScalingEToD65 is the matrix for conversion from illuminant E to D65.
var XYZScalingEToD65 = [...]float32{
	0.9504700, 0.0000000, 0.0000000,
	0.0000000, 1.0000000, 0.0000000,
	0.0000000, 0.0000000, 1.0888300,
}

// ChromaticAdjust shifts the white-point of the XYZ colour using one of the Bradford* or XYZScaling* matrices.
func ChromaticAdjust(mat [9]float32, a [3]float32) [3]float32 {
	var b [3]float32

	b[0] = a[0]*mat[0] + a[1]*mat[1] + a[2]*mat[2]
	b[1] = a[0]*mat[3] + a[1]*mat[4] + a[2]*mat[5]
	b[2] = a[0]*mat[6] + a[1]*mat[7] + a[2]*mat[8]

	return b
}
