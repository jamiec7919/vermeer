// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	The math package implements 3D maths operations.

	Vermeer relies on at least SSE with support for SQRT instructions and no provision
	is made for other platforms than AMD64 yet (will fail to compile).
*/
package math

// Vec3 represents a 3 dimensional vector (32 bit floats)
type Vec3 [3]float32

// Vec3Mad computes a mul-and-add operation, a + b * s
func Vec3Mad(a, b Vec3, s float32) (v Vec3) {
	v[0] = a[0] + (b[0] * s)
	v[1] = a[1] + (b[1] * s)
	v[2] = a[2] + (b[2] * s)
	return
}

func Vec3Add(a, b Vec3) (v Vec3) {
	v[0] = a[0] + b[0]
	v[1] = a[1] + b[1]
	v[2] = a[2] + b[2]

	return
}

func Vec3Add3(a, b, c Vec3) (v Vec3) {
	v[0] = a[0] + b[0] + c[0]
	v[1] = a[1] + b[1] + c[1]
	v[2] = a[2] + b[2] + c[2]

	return
}

func Vec3Sub(a, b Vec3) (v Vec3) {
	v[0] = a[0] - b[0]
	v[1] = a[1] - b[1]
	v[2] = a[2] - b[2]

	return
}

func Vec3Lerp(a, b Vec3, t float32) (v Vec3) {
	v[0] = (1.0-t)*a[0] + t*b[0]
	v[1] = (1.0-t)*a[1] + t*b[1]
	v[2] = (1.0-t)*a[2] + t*b[2]

	return
}

func Vec3Scale(s float32, a Vec3) (v Vec3) {
	v[0] = a[0] * s
	v[1] = a[1] * s
	v[2] = a[2] * s
	return
}

func Vec3Cross(a, b Vec3) (v Vec3) {
	v[0] = a[1]*b[2] - a[2]*b[1]
	v[1] = a[2]*b[0] - a[0]*b[2]
	v[2] = a[0]*b[1] - a[1]*b[0]
	return
}

func Vec3Dot(a, b Vec3) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func Vec3DotClamp(a, b Vec3) float32 {
	return Max(0.0, a[0]*b[0]+a[1]*b[1]+a[2]*b[2])
}

func Vec3DotAbs(a, b Vec3) float32 {
	return Abs(a[0]*b[0] + a[1]*b[1] + a[2]*b[2])
}

func Vec3Normalize(a Vec3) (v Vec3) {
	d := Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])

	v = Vec3Scale(1.0/d, a)
	return
}

func Vec3Length2(a Vec3) float32 {
	return a[0]*a[0] + a[1]*a[1] + a[2]*a[2]
}

func Vec3Length(a Vec3) float32 {
	return Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
}

func Vec3Neg(a Vec3) (v Vec3) {
	v[0] = -a[0]
	v[1] = -a[1]
	v[2] = -a[2]
	return
}

func Vec3BasisProject(U, V, W, S Vec3) (o Vec3) {
	o[0] = Vec3Dot(U, S)
	o[1] = Vec3Dot(V, S)
	o[2] = Vec3Dot(W, S)
	return
}

// Vec3Basis calculates the vector S in the basis defined by U,V,W.
// O := U*S_x + V*S_y
func Vec3BasisExpand(U, V, W, S Vec3) (o Vec3) {
	o[0] = U[0]*S[0] + V[0]*S[1] + W[0]*S[2]
	o[1] = U[1]*S[0] + V[1]*S[1] + W[1]*S[2]
	o[2] = U[2]*S[0] + V[2]*S[1] + W[2]*S[2]

	return
}

func Vec3Abs(V Vec3) (o Vec3) {
	o[0] = Abs(V[0])
	o[1] = Abs(V[1])
	o[2] = Abs(V[2])
	return
}

func (v Vec3) MaxDim() int {
	if v[0] < v[1] {
		if v[1] < v[2] {
			return 2
		} else {
			return 1
		}
	} else {
		if v[0] < v[2] {
			return 2
		} else {
			return 0
		}

	}
}
