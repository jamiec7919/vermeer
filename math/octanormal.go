// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/*
The OctaNormal type represents octahedral normals.  Converts a 3 component direction vector (doesn't have to be
normalized as this is done on encoding) to a 2 component octahedral normal vector.

Note that the reference paper allows for smaller lossy encodings but this just implements
the lossless conversion to two float32s.
*/
type OctaNormal [2]float32

func octWrap(v OctaNormal) (on OctaNormal) {
	on[0] = (1.0 - Abs(v[1])) * sign(v[0])
	on[1] = (1.0 - Abs(v[0])) * sign(v[1])

	return
}

// EncodeOctahedralNormal converts the (unit) vector n to the Octahedral normal representation.
func EncodeOctahedralNormal(n Vec3) (on OctaNormal) {
	nf := Abs(n.X) + Abs(n.Y) + Abs(n.Z)

	n.X /= nf
	n.Y /= nf
	n.Z /= nf

	if n.Z >= 0.0 {
		on[0] = n.X
		on[1] = n.Y
	} else {
		on = octWrap(OctaNormal{n.X, n.Y})

	}

	on[0] = on[0]*0.5 + 0.5
	on[1] = on[1]*0.5 + 0.5
	return
}

// DecodeOctahedralNormal converts the octahedral normal into a unit vector.
func DecodeOctahedralNormal(on OctaNormal) Vec3 {
	on[0] = on[0]*2.0 - 1.0
	on[1] = on[1]*2.0 - 1.0

	var n Vec3
	n.Z = 1.0 - Abs(on[0]) - Abs(on[1])
	if n.Z >= 0.0 {
		n.X = on[0]
		n.Y = on[1]
	} else {
		o := octWrap(on)
		n.X = o[0]
		n.Y = o[1]
	}
	return Vec3Normalize(n)
}
