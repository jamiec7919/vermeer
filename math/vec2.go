// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

type Vec2 [2]float32

func Vec2Add(a, b Vec2) (v Vec2) {
	v[0] = a[0] + b[0]
	v[1] = a[1] + b[1]

	return
}

func Vec2Sub(a, b Vec2) (v Vec2) {
	v[0] = a[0] - b[0]
	v[1] = a[1] - b[1]

	return
}

func Vec2Scale(s float32, a Vec2) (v Vec2) {
	v[0] = a[0] * s
	v[1] = a[1] * s

	return
}

func Vec2Dot(a, b Vec2) float32 {
	return a[0]*b[0] + a[1]*b[1]
}

func Vec2Length2(a Vec2) float32 {
	return a[0]*a[0] + a[1]*a[1]
}

func Vec2Length(a Vec2) float32 {
	return Sqrt(a[0]*a[0] + a[1]*a[1])
}

func Vec2Mad(a, b Vec2, s float32) (v Vec2) {
	v[0] = a[0] + (b[0] * s)
	v[1] = a[1] + (b[1] * s)
	return
}
