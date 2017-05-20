// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package colour provides spectral, hero-wavelength and RGB support functions and types.
package colour

// Space represents a transform from XYZ to a different space.
type Space struct {
	m    [9]float32 // Matrix to convert XYZ->RGB
	mInv [9]float32 // Matrix to convert XYZ->RGB
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
