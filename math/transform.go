// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/* Transform represents an interpolatable transformation consisting of a translation,
rotation and a scale factor.

Idea and algorithm from: http://research.cs.wisc.edu/graphics/Courses/838-s2002/Papers/polar-decomp.pdf

Supplementary notes: http://www.cs.cornell.edu/courses/cs4620/2014fa/lectures/polarnotes.pdf
*/
type Transform struct {
	Translate Vec3
	Rotate    Quat
	Scale     Matrix4
}

// Matrix4ToTransform decomposes the given matrix into components.
func Matrix4ToTransform(m Matrix4) Transform {
	var decomp Transform

	//TODO: Add projection
	//TODO: if m has det < 0 then should also factor out -I beforehand.
	sign := sign(Matrix4Det(m))

	// Extract transform component and remove from m.
	decomp.Translate.X = m.Elt(0, 3)
	decomp.Translate.Y = m.Elt(1, 3)
	decomp.Translate.Z = m.Elt(2, 3)

	m.Set(0, 3, 0)
	m.Set(1, 3, 0)
	m.Set(2, 3, 0)

	if sign < 0.0 {
		m = Matrix4Mul(Matrix4Scale(-1, Matrix4Identity()), m)
	}

	Q, ok := Matrix4PolarFactor(m)

	if !ok {
		// Something bad happened, just bail.
		decomp.Rotate = QuatIdentity()
		decomp.Scale = Matrix4Identity()
		return decomp
	}

	// S = Q^T M
	S := Matrix4Mul(Matrix4Transpose(Q), m)

	decomp.Rotate = Matrix4ToQuat(Q)

	if sign < 0.0 {
		decomp.Scale = Matrix4Mul(Matrix4Scale(-1, Matrix4Identity()), S)
		decomp.Scale[15] = 1 // Factor of -I will only affect upper 3x3 since quaternion representation
		// will only affect that portion of matrix.  We manually set the homogenous
		// factor here.
	} else {

		decomp.Scale = S
	}

	return decomp
}

// TransformToMatrix4 composes the given transform components into a matrix.
func TransformToMatrix4(decomp Transform) Matrix4 {

	return Matrix4Mul(Matrix4TransformTranslate(decomp.Translate.X, decomp.Translate.Y, decomp.Translate.Z),
		Matrix4Mul(QuatToMatrix4(decomp.Rotate), decomp.Scale))

}

// TransformLerp linearly interpolates between two transforms.
func TransformLerp(a, b Transform, t float32) Transform {
	var out Transform

	out.Translate = Vec3Lerp(a.Translate, b.Translate, t)
	out.Rotate = QuatSlerp(a.Rotate, b.Rotate, t)
	out.Scale = Matrix4Lerp(a.Scale, b.Scale, t)

	return out
}
