// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// TransformDecomp is a decomposed Matrix4 into interpolable components. Currently we assume
// that there is no projection.
//
// Idea and algorithm from: http://research.cs.wisc.edu/graphics/Courses/838-s2002/Papers/polar-decomp.pdf
//
// Supplementary notes: http://www.cs.cornell.edu/courses/cs4620/2014fa/lectures/polarnotes.pdf
type TransformDecomp struct {
	T Vec3
	R Quat
	S Matrix4
}

// TransformDecompMatrix4 decomposes the given matrix into components.
func TransformDecompMatrix4(m Matrix4) (decomp TransformDecomp) {
	//TODO: Add projection
	//TODO: if m has det < 0 then should also factor out -I beforehand.
	sign := sign(Matrix4Det(m))

	// Extract transform component and remove from m.
	for i := range decomp.T {
		decomp.T[i] = m.Elt(i, 3)
	}

	m.Set(0, 3, 0)
	m.Set(1, 3, 0)
	m.Set(2, 3, 0)

	if sign < 0.0 {
		m = Matrix4Mul(Matrix4Scale(-1, Matrix4Identity), m)
	}

	Q, ok := Matrix4PolarFactor(m)

	if !ok {
		// Something bad happened, just bail.
		decomp.R = QuatIdentity
		decomp.S = Matrix4Identity
		return
	}

	// S = Q^T M
	S := Matrix4Mul(Matrix4Transpose(Q), m)

	decomp.R = Matrix4ToQuat(Q)

	if sign < 0.0 {
		decomp.S = Matrix4Mul(Matrix4Scale(-1, Matrix4Identity), S)
		decomp.S[15] = 1 // Factor of -I will only affect upper 3x3 since quaternion representation
		// will only affect that portion of matrix.  We manually set the homogenous
		// factor here.
	} else {

		decomp.S = S
	}

	return
}

// TransformDecompToMatrix4 composes the given transform components into a matrix.
func TransformDecompToMatrix4(decomp TransformDecomp) (m Matrix4) {
	m = Matrix4Mul(Matrix4Translate(decomp.T[0], decomp.T[1], decomp.T[2]),
		Matrix4Mul(QuatToMatrix4(decomp.R), decomp.S))

	return
}

// TransformDecompLerp linearly interpolates between two transforms.
func TransformDecompLerp(a, b TransformDecomp, t float32) (out TransformDecomp) {
	out.T = Vec3Lerp(a.T, b.T, t)
	out.R = QuatSlerp(a.R, b.R, t)
	out.S = Matrix4Lerp(a.S, b.S, t)

	return
}
