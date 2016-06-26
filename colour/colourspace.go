// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package colour provides spectral, hero-wavelength and RGB support functions and types.
package colour

// Space represents a transform from XYZ to a different space.
type Space struct {
	m [9]float32 // Matrix to convert XYZ->RGB
}

// XYZToRGB returns the transformed XYZ to RGB
func (s *Space) XYZToRGB(x, y, z float32) (r, g, b float32) {
	r = x*s.m[0] + y*s.m[1] + z*s.m[2]
	g = x*s.m[3] + y*s.m[4] + z*s.m[5]
	b = x*s.m[6] + y*s.m[7] + z*s.m[8]

	return
}
