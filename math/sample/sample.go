// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sample

import (
	m "github.com/jamiec7919/vermeer/math"
	"math"
)

// pdf?  Think it is cos(v,N)/Pi
func CosineHemisphere(u0, u1 float64) m.Vec3 {
	r := math.Sqrt(u0)
	theta := 2 * math.Pi * u1

	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(math.Sqrt(1 - u0))}

}

func UniformSphere(u0, u1 float64) m.Vec3 {
	theta := 2 * u0 * math.Pi
	z := -1.0 + 2*u1

	r := math.Sqrt(1 - z*z)
	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(z)}
}

func UniformDisk2D(radius, r0, r1 float32) (xo, yo float32) {
	// Square to disk warp

	x := -1 + 2*r0
	y := -1 + 2*r1
	r, theta := float32(0), float32(0)
	if x > -y && x > y {
		r = x
		theta = (m.Pi / 4) * y / x
	} else if x > -y && x < y {
		r = y
		theta = (m.Pi / 4) * (2 - x/y)

	} else if x < y && x < -y {
		r = -x
		theta = (m.Pi / 4) * (4 + y/x)

	} else if x > y && x < -y {
		r = -y
		theta = (m.Pi / 4) * (6 - x/y)
	}
	xo = radius * r * m.Cos(theta)
	yo = radius * r * m.Sin(theta)
	return
}
