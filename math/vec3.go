// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"math"
)

// Vec3 represents a 3D vector.
type Vec3 struct {
	X, Y, Z float32
}

// Vec3Mad computes a mul-and-add operation, a + b * s
func Vec3Mad(a, b Vec3, s float32) Vec3 {
	var v Vec3

	v.X = a.X + (b.X * s)
	v.Y = a.Y + (b.Y * s)
	v.Z = a.Z + (b.Z * s)
	return v
}

// Vec3Add returns the sum of 3 dimension vectors a and b.
func Vec3Add(a, b Vec3) Vec3 {
	var v Vec3

	v.X = a.X + b.X
	v.Y = a.Y + b.Y
	v.Z = a.Z + b.Z

	return v
}

// Vec3Add3 adds 3 3-vectors a,b and c.
func Vec3Add3(a, b, c Vec3) Vec3 {
	var v Vec3

	v.X = a.X + b.X + c.X
	v.Y = a.Y + b.Y + c.Y
	v.Z = a.Z + b.Z + c.Z

	return v
}

// Vec3Sub returns the difference of 3 dimensional vectors a and b.
func Vec3Sub(a, b Vec3) Vec3 {
	var v Vec3

	v.X = a.X - b.X
	v.Y = a.Y - b.Y
	v.Z = a.Z - b.Z

	return v
}

// Vec3Lerp linearly interpolates from a to b based on parameter t in [0,1].
func Vec3Lerp(a, b Vec3, t float32) Vec3 {
	var v Vec3

	v.X = (1.0-t)*a.X + t*b.X
	v.Y = (1.0-t)*a.Y + t*b.Y
	v.Z = (1.0-t)*a.Z + t*b.Z

	return v
}

// Vec3Scale returns the 3-vector a scaled by s.
func Vec3Scale(s float32, a Vec3) Vec3 {
	var v Vec3

	v.X = a.X * s
	v.Y = a.Y * s
	v.Z = a.Z * s

	return v
}

// Vec3Cross returns the 3D cross-product of a and b.
func Vec3Cross(a, b Vec3) Vec3 {
	var v Vec3

	v.X = a.Y*b.Z - a.Z*b.Y
	v.Y = a.Z*b.X - a.X*b.Z
	v.Z = a.X*b.Y - a.Y*b.X

	return v
}

// Vec3Dot returns the 3D dot product of a and b.
func Vec3Dot(a, b Vec3) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// Vec3DotClamp returns the 3D dot product of a and b clamped to [0,+inf].
func Vec3DotClamp(a, b Vec3) float32 {
	return Max(0.0, a.X*b.X+a.Y*b.Y+a.Z*b.Z)
}

// Vec3DotAbs returns the absolute value of the 3D dot product of a and b.
func Vec3DotAbs(a, b Vec3) float32 {
	return Abs(a.X*b.X + a.Y*b.Y + a.Z*b.Z)
}

// Vec3NormalizeRSqrt returns the unit vector in the same direction as a.
// Uses reciprocal-sqrt instruction plus Newton-Raphson step.  Not really accurate
// enough but might be faster occasionally.
func Vec3NormalizeRSqrt(a Vec3) Vec3

func Vec3Normalize(a Vec3) Vec3

// Vec3Normalize returns the unit vector in the same direction as a.
func _Vec3Normalize(a Vec3) Vec3 {
	var v Vec3

	d := float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z)))

	//v = Vec3Scale(1.0/d, a)
	v.X = a.X / d
	v.Y = a.Y / d
	v.Z = a.Z / d

	return v
}

// Vec3Length2 returns the squared length of vector a.
func Vec3Length2(a Vec3) float32 {
	return a.X*a.X + a.Y*a.Y + a.Z*a.Z
}

// Vec3Length returns the length of vector a.
func Vec3Length(a Vec3) float32

func vec3Length(a Vec3) float32 {
	return Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}

// Vec3Neg returns the negative of vector a.
func Vec3Neg(a Vec3) Vec3 {
	var v Vec3
	v.X = -a.X
	v.Y = -a.Y
	v.Z = -a.Z
	return v
}

// Vec3BasisProject returns the vector S projected onto the basis U,V,W.
func Vec3BasisProject(U, V, W, S Vec3) Vec3 {
	var o Vec3

	o.X = Vec3Dot(U, S)
	o.Y = Vec3Dot(V, S)
	o.Z = Vec3Dot(W, S)
	return o
}

// Vec3BasisExpand calculates the vector S in the basis defined by U,V,W.
// O := U*S_x + V*S_y + W*S_z
func Vec3BasisExpand(U, V, W, S Vec3) Vec3 {
	var o Vec3

	o.X = U.X*S.X + V.X*S.Y + W.X*S.Z
	o.Y = U.Y*S.X + V.Y*S.Y + W.Y*S.Z
	o.Z = U.Z*S.X + V.Z*S.Y + W.Z*S.Z

	return o
}

// Vec3Abs returns the component-wise absolute of V.
func Vec3Abs(V Vec3) Vec3 {
	var o Vec3

	o.X = Abs(V.X)
	o.Y = Abs(V.Y)
	o.Z = Abs(V.Z)

	return o
}

// MaxDim returns the axis in which the vector is maximal.
func (v Vec3) MaxDim() int {
	if v.X < v.Y {
		if v.Y < v.Z {
			return 2

		}

		return 1

	}

	if v.X < v.Z {
		return 2
	}

	return 0

}

/** These are implemented in vec3_unsafe.go
// Elt returns the i'th element of the vector.
func (v *Vec3) Elt(i int) float32 {
	switch i {
	case 0:
		return a.X
	case 1:
		return a.Y
	case 2:
		return a.Z
	}

	panic("Index out of range.")

}

// Set sets the value of the i'th component of the vector.
func (v *Vec3) Set(i int, a float32) {
	switch i {
	case 0:
		a.X = a
	case 1:
		a.Y = a
	case 2:
		a.Z = a
	}

	panic("Index out of range.")

}
*/
