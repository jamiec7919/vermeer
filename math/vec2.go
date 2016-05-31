// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Vec2 represents a 2 dimensional vector (32 bit floats)
type Vec2 [2]float32

// Vec2Add returns the sum of 2 dimension vectors a and b.
func Vec2Add(a, b Vec2) (v Vec2) {
	v[0] = a[0] + b[0]
	v[1] = a[1] + b[1]

	return
}

// Vec2Sub returns the difference of 2 dimension vectors a and b.
func Vec2Sub(a, b Vec2) (v Vec2) {
	v[0] = a[0] - b[0]
	v[1] = a[1] - b[1]

	return
}

// Vec2Scale returns the 2-vector a scaled by s.
func Vec2Scale(s float32, a Vec2) (v Vec2) {
	v[0] = a[0] * s
	v[1] = a[1] * s

	return
}

// Vec2Lerp linearly interpolates from a to b based on parameter t in [0,1].
func Vec2Lerp(a, b Vec2, t float32) (v Vec2) {
	v[0] = (1.0-t)*a[0] + t*b[0]
	v[1] = (1.0-t)*a[1] + t*b[1]

	return
}

// Vec2Dot computes the 2D dot product of vectors a and b.
func Vec2Dot(a, b Vec2) float32 {
	return a[0]*b[0] + a[1]*b[1]
}

// Vec2Length2 returns the squared length of vector a.
func Vec2Length2(a Vec2) float32 {
	return a[0]*a[0] + a[1]*a[1]
}

// Vec2Length returns the length of vector a.
func Vec2Length(a Vec2) float32 {
	return Sqrt(a[0]*a[0] + a[1]*a[1])
}

// Vec2Mad returns the multiply-add:  a + s*b
func Vec2Mad(a, b Vec2, s float32) (v Vec2) {
	v[0] = a[0] + (b[0] * s)
	v[1] = a[1] + (b[1] * s)
	return
}
