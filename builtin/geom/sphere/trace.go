// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sphere

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

// Trace implements core.Primitive.
func (sphere *Sphere) Trace(ray *core.Ray, sg *core.ShaderContext) bool {

	t, ok := raySphereIntersect(ray.P, ray.D, sphere.P, sphere.Radius)

	if ok {
		if t < ray.Tclosest && t > ray.Tmin {
			ray.Tclosest = t

			P := m.Vec3Mad(ray.P, ray.D, t)

			N := m.Vec3Normalize(m.Vec3Sub(P, sphere.P))

			U := 0.5 + m.Atan2(N.Z, N.X)/(2*m.Pi)
			V := 0.5 - m.Asin(N.Y)/m.Pi

			sg.Poffset = m.Vec3Scale(0.001, N)
			sg.Ng = N
			sg.N = N
			sg.Bu = U
			sg.Bv = V
			sg.P = P
			sg.Po = sg.P
			sg.U = U
			sg.V = V

			sg.DdPdu.X = 1
			sg.DdPdu.Y = 0
			sg.DdPdu.Z = 0

			sg.DdPdv.X = 0
			sg.DdPdv.Y = 0
			sg.DdPdv.Z = 1

			sg.Shader = sphere.shader

			return true
		}
	}
	return false
}

func solveQuadratic(a, b, c float32, x0, x1 *float32) bool {
	discr := b*b - 4*a*c

	if discr < 0 {
		return false
	} else if discr == 0 {
		*x1 = -0.5 * b / a
		*x0 = *x1
	} else {
		var q float32

		if b > 0 {
			q = -0.5 * (b + m.Sqrt(discr))
		} else {
			q = -0.5 * (b - m.Sqrt(discr))
		}
		*x0 = q / a
		*x1 = c / q
	}

	if *x0 > *x1 {
		*x0, *x1 = *x1, *x0
	}

	return true
}

func raySphereIntersect(Ro, Rd, P m.Vec3, radius float32) (float32, bool) {
	// analytic solution
	L := m.Vec3Sub(Ro, P)

	a := m.Vec3Dot(Rd, Rd)

	b := 2 * m.Vec3Dot(Rd, L)

	c := m.Vec3Dot(L, L) - radius*radius

	var t0, t1 float32

	if !solveQuadratic(a, b, c, &t0, &t1) {
		return 0, false
	}

	if t0 > t1 {
		t0, t1 = t1, t0
	}

	if t0 < 0 {
		t0 = t1 // if t0 is negative, let's use t1 instead
		if t0 < 0 {
			return 0, false
		} // both t0 and t1 are negative
	}

	return t0, true
}
