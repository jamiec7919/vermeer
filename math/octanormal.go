// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/*
	Octahedral normals.  Converts a 3 component direction vector (doesn't have to be
	normalized as this is done on encoding) to a 2 component octahedral normal vector.

	Note that the reference paper allows for smaller lossy encodings but this just implements
	the lossless conversion to two float32s.
*/

type OctaNormal [2]float32

func sign(v float32) float32 {
	if v >= 0.0 {
		return 1.0
	} else {
		return -1.0
	}

}

func octWrap(v OctaNormal) (on OctaNormal) {
	on[0] = (1.0 - Abs(v[1])) * sign(v[0])
	on[1] = (1.0 - Abs(v[0])) * sign(v[1])

	return
}

func EncodeOctahedralNormal(n Vec3) (on OctaNormal) {
	nf := Abs(n[0]) + Abs(n[1]) + Abs(n[2])

	n[0] /= nf
	n[1] /= nf
	n[2] /= nf

	if n[2] >= 0.0 {
		on[0] = n[0]
		on[1] = n[1]
	} else {
		on = octWrap(OctaNormal{n[0], n[1]})

	}

	on[0] = on[0]*0.5 + 0.5
	on[1] = on[1]*0.5 + 0.5
	return
}

func DecodeOctahedralNormal(on OctaNormal) Vec3 {
	on[0] = on[0]*2.0 - 1.0
	on[1] = on[1]*2.0 - 1.0

	var n Vec3
	n[2] = 1.0 - Abs(on[0]) - Abs(on[1])
	if n[2] >= 0.0 {
		n[0] = on[0]
		n[1] = on[1]
	} else {
		o := octWrap(on)
		n[0] = o[0]
		n[1] = o[1]
	}
	return Vec3Normalize(n)
}
