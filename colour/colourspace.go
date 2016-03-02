// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

type RGB struct {
	R, G, B float32
}

type ColourSpace struct {
	m [9]float32 // Matrix to convert XYZ->RGB
}

//Transform XYZ to RGB
func (s *ColourSpace) XYZToRGB(x, y, z float32) (r, g, b float32) {
	r = x*s.m[0] + y*s.m[1] + z*s.m[2]
	g = x*s.m[3] + y*s.m[4] + z*s.m[5]
	b = x*s.m[6] + y*s.m[7] + z*s.m[8]

	return
}
