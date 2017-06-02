// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Vec3ProjectPerp projects v perpendicular to the vector n.
func Vec3ProjectPerp(v, n Vec3) Vec3 {
	return Vec3Sub(v, Vec3Scale(Vec3Dot(v, n), n))
}

// Vec3ProjectParallel projects v parallel to the vector n.
func Vec3ProjectParallel(v, n Vec3) Vec3 {
	return Vec3Scale(Vec3Dot(v, n), n)
}

// Vec3Reflect reflects v about vector n. This assumes that v is heading into the plane
// defined by N and will return the vector pointing away from the plane.  The same is
// true on the other side of the plane.
func Vec3Reflect(v, N Vec3) Vec3 {
	return Vec3Add(v, Vec3Scale(2.0*Vec3Dot(v, N), N))
}

// Vec3Refract implements optical refraction at an interface between dielectrics. Assumes vector
// is heading into the plane.
// ior = n/n_t  where n is index of refraction of the 'from' medium and n_t is the index of refraction
// of the 'to' medium.  Returns false and reflected direction if total internal reflection occurs.
func Vec3Refract(d, N Vec3, ior float32) (Vec3, bool) {
	// ior = n/n_t or inverse
	dotN := Vec3Dot(d, N)

	sq := 1 - ior*ior*(1-dotN*dotN)

	// Total internal reflection
	if sq < 0 {
		return Vec3Reflect(d, N), false
	}

	omega := Vec3Sub(Vec3Scale(ior, Vec3Sub(d, Vec3Scale(dotN, N))), Vec3Scale(Sqrt(sq), N))

	return omega, true
}
