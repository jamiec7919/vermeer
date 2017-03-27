// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package sample provides various sampling strategies and utilities.
*/
package sample

import (
	m "github.com/jamiec7919/vermeer/math"
	"math"
)

// CosineHemisphere returns a unit vector sampled from the cosine weighted hemisphere. Normal
// is [0,0,1]
// pdf is cos(v,N)/Pi
func CosineHemisphere(u0, u1 float64) m.Vec3 {
	r := math.Sqrt(1 - u0)
	theta := 2 * math.Pi * u1

	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(math.Sqrt(u0))}

}

// CosineHemisphere2 returns a unit vector sampled from the cosine weighted hemisphere. Normal
// is [0,0,1]
// pdf is cos(v,N)/Pi
func CosineHemisphere2(u0, u1 float64) m.Vec3 {
	r := math.Sqrt(u0)
	theta := 2 * math.Pi * u1

	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(math.Sqrt(1 - u0))}

}

// CosineHemisphereConcentric returns a unit vector sampled from the cosine weighted hemisphere. Normal
// is [0,0,1].  Uses Concentric (Shirley) mapping.
// pdf is cos(v,N)/Pi
func CosineHemisphereConcentric(u0, u1 float64) m.Vec3 {
	var r, phi float64

	u0 = -1 + (u0 * 2)
	u1 = -1 + (u1 * 2)

	switch {
	case u0 > -u1 && u0 > u1:
		r = u0
		phi = (math.Pi / 4) * (u1 / u0)
	case u0 < u1 && u0 > -u1:
		r = u1
		phi = (math.Pi / 4) * (2 - u0/u1)
	case u0 < -u1 && u0 < u1:
		r = -u0
		phi = (math.Pi / 4) * (4 + u1/u0)
	case u0 > u1 && u0 < -u1:
		r = -u1
		phi = (math.Pi / 4) * (6 - u0/u1)

	}

	x := r * math.Cos(phi)
	y := r * math.Sin(phi)

	return m.Vec3{float32(x), float32(y), m.Sqrt(1 - float32(r*r))}

}

// UniformHemisphere returns a unit vector sampled from the cosine weighted hemisphere. Normal
// is [0,0,1]
// pdf is 1/2pi
func UniformHemisphere(u0, u1 float64) m.Vec3 {
	r := math.Sqrt(1 - u1*u1)
	theta := 2 * math.Pi * u0

	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(u1)}

}

// UniformSphere returns a unit vector uniformly sampled from the sphere.
// pdf is 1/(4*Pi)
func UniformSphere(u0, u1 float64) m.Vec3 {
	theta := 2 * u0 * math.Pi
	z := -1.0 + 2*u1

	r := math.Sqrt(1 - z*z)
	x := r * math.Cos(theta)
	y := r * math.Sin(theta)

	return m.Vec3{float32(x), float32(y), float32(z)}
}

// UniformDisk2D returns a 2D point sampled from the disk with given radius.  Uses
// uniform warping.
// pdf is 1/Pi*radius^2
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
