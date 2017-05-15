// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/* Matrix4 represents a 4x4 homogenous matrix.

Column major

[ 0 4 8 12 ]
[ 1 5 9 13 ]
[ 2 6 10 14]
[ 3 7 11 15]

Transforms must POST multiply matrix by column vector.
*/
type Matrix4 [16]float32

// IsNull returns true if matrix is the 0 matrix (all elements 0).
func (m *Matrix4) IsNull() bool {
	for i := range m {
		if m[i] != 0.0 && m[i] != -0.0 {
			return false
		}
	}
	return true
}

// IsIdentity returns true if matrix is the identity matrix.
func (m *Matrix4) IsIdentity() bool {
	for i := range m {
		switch i {
		case 0, 5, 10, 15:
			if m[i] != 1.0 {
				return false
			}
		default:
			if m[i] != 0.0 && m[i] != -0.0 {
				return false
			}
		}
	}
	return true
}

// Elt returns the matrix element at row i, column j.
func (m *Matrix4) Elt(i, j int) float32 {
	return m[(j*4)+i]
}

// Set sets the matrix element at row i, column j.
func (m *Matrix4) Set(i, j int, v float32) {
	m[(j*4)+i] = v
}

// Matrix4Identity is the identity matrix.
func Matrix4Identity() Matrix4 {
	return Matrix4{
		1.0, 0.0, 0.0, 0.0,
		0.0, 1.0, 0.0, 0.0,
		0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 1.0}
}

// Matrix4Null is the zero matrix.
func Matrix4Null() Matrix4 {
	return Matrix4{
		0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0}
}

// Matrix4Add returns the sum of 4x4 matrices a and b.
func Matrix4Add(a, b Matrix4) Matrix4 {
	var c Matrix4

	for i := range c {
		c[i] = a[i] + b[i]
	}

	return c
}

// Matrix4Sub returns the difference of 4x4 matrices a and b.
func Matrix4Sub(a, b Matrix4) Matrix4 {
	var c Matrix4

	for i := range c {
		c[i] = a[i] - b[i]
	}

	return c
}

// Matrix4Mul returns the matrix multiplication of 4x4 matrices a and b.
func Matrix4Mul(a, b Matrix4) Matrix4 {
	var c Matrix4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				c[(j*4)+i] += a[(k*4)+i] * b[(j*4)+k]
			}
		}
	}

	return c
}

// Matrix4Transpose returns the transpose of 4x4 matrix a.
func Matrix4Transpose(a Matrix4) Matrix4 {
	var c Matrix4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			c[(i*4)+j] = a[(j*4)+i]
		}
	}

	return c
}

// Matrix4Inverse returns the matrix inverse of 4x4 matrix a.
func Matrix4Inverse(m Matrix4) (Matrix4, bool) {
	var inv Matrix4

	inv[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] + m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	inv[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] - m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	inv[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] + m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	inv[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] - m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]
	inv[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] - m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	inv[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] + m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	inv[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] - m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	inv[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] + m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]
	inv[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] + m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	inv[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] - m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	inv[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] + m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	inv[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] - m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]
	inv[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] - m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	inv[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] + m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	inv[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] - m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	inv[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] + m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	det := m[0]*inv[0] + m[1]*inv[4] + m[2]*inv[8] + m[3]*inv[12]

	if det == 0.0 {
		return inv, false
	}

	det = 1.0 / det

	for i := range inv {
		inv[i] *= det
	}

	return inv, true
}

// Matrix4Det returns the determinant of 4x4 matrix a.
func Matrix4Det(m Matrix4) float32 {
	var inv Matrix4

	inv[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] + m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	inv[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] - m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	inv[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] + m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	inv[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] - m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]
	inv[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] - m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	inv[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] + m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	inv[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] - m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	inv[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] + m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]
	inv[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] + m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	inv[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] - m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	inv[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] + m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	inv[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] - m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]
	inv[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] - m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	inv[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] + m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	inv[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] - m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	inv[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] + m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	return m[0]*inv[0] + m[1]*inv[4] + m[2]*inv[8] + m[3]*inv[12]
}

// Matrix4MulPoint post multiplies point b by 4x4 matrix a.
// b is represented as a Vec3 but is assumed to be homogeonous point [x,y,z,1]
func Matrix4MulPoint(a Matrix4, b Vec3) Vec3 {
	var c Vec3

	c.X = a[0]*b.X + a[4]*b.Y + a[8]*b.Z + a[12]
	c.Y = a[1]*b.X + a[5]*b.Y + a[9]*b.Z + a[13]
	c.Z = a[2]*b.X + a[6]*b.Y + a[10]*b.Z + a[14]

	return c
}

// Matrix4MulHPoint post multiplies point b by 4x4 matrix a.
// b is represented as a full homoegenous point and returns in same representation.
func Matrix4MulHPoint(a Matrix4, b [4]float32) [4]float32 {
	var c [4]float32

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			c[i] += a[(j*4)+i] * b[j]
		}
	}

	return c
}

// Matrix4MulVec post multiplies vector b by 4x4 matrix a.
// b is represented as a Vec3 but is assumed to be homogeonous vector [x,y,z,0]
func Matrix4MulVec(a Matrix4, b Vec3) Vec3 {
	var c Vec3

	c.X = a[0]*b.X + a[4]*b.Y + a[8]*b.Z
	c.Y = a[1]*b.X + a[5]*b.Y + a[9]*b.Z
	c.Z = a[2]*b.X + a[6]*b.Y + a[10]*b.Z

	return c
}

// Matrix4TransformTranslate returns the matrix representing a translation by the given amount.
func Matrix4TransformTranslate(X, Y, Z float32) Matrix4 {
	var c Matrix4
	c[0] = 1
	c[5] = 1
	c[10] = 1
	c[12] = X
	c[13] = Y
	c[14] = Z
	c[15] = 1
	return c
}

// Matrix4TransformScale returns the matrix representing a scaling by the given amount.
func Matrix4TransformScale(X, Y, Z float32) Matrix4 {
	var c Matrix4
	c[0] = X
	c[5] = Y
	c[10] = Z
	c[15] = 1
	return c
}

// Matrix4Scale returns the matrix m multiplied by scalar s.
func Matrix4Scale(s float32, m Matrix4) Matrix4 {
	var c Matrix4

	for i := range m {
		c[i] = s * m[i]
	}

	return c
}

// Matrix4Lerp returns the linear interpolation between matrix a and b. This
// is only useful in very restricted circumstances.
func Matrix4Lerp(a, b Matrix4, t float32) Matrix4 {
	var c Matrix4

	for i := range c {
		c[i] = (1.0-t)*a[i] + t*b[i]
	}

	return c
}

// Matrix4TransformRotate returns the matrix representing a rotation by angle (radians) around axis [X,Y,Z].
/*func Matrix4TransformRotate(angle, X, Y, Z float32) Matrix4 {
	var c Matrix4

	// Should normalize X,Y,Z
	sl := X*X + Y*Y + Z*Z

	if sl < 1.0-0.00001 || sl > 1.0+0.00001 {
		sl2 := Sqrt(sl)
		X = X / sl2
		Y = Y / sl2
		Z = Z / sl2

	}

	C := Cos(angle)
	S := Sin(angle)

	c[0] = X*X*(1-C) + C
	c[4] = X*Y*(1-C) - Z*S
	c[8] = X*Z*(1-C) + Y*S
	c[12] = 0.0

	c[1] = Y*X*(1-C) + Z*S
	c[5] = Y*Y*(1-C) + C
	c[9] = Y*Z*(1-C) - X*S
	c[13] = 0.0

	c[2] = Z*X*(1-C) - Y*S
	c[6] = Z*Y*(1-C) + X*S
	c[10] = Z*Z*(1-C) + C
	c[14] = 0.0

	c[3] = 0.0
	c[7] = 0.0
	c[11] = 0.0
	c[15] = 1.0

	return c
}*/

// Matrix4Basis constructs a 4x4 matrix from three basis vectors (for the 3x3 bit)
func Matrix4Basis(u, v, w Vec3) Matrix4 {
	var out Matrix4

	out[0] = u.X
	out[1] = u.Y
	out[2] = u.Z

	out[4] = v.X
	out[5] = v.Y
	out[6] = v.Z

	out[8] = w.X
	out[9] = w.Y
	out[10] = w.Z

	out[15] = 1.0

	return out
}

// Matrix4Eq returns true if A-B approx= 0, within given epsilon.
func Matrix4Eq(A, B Matrix4, eps float32) bool {
	for i := range A {
		if A[i]-B[i] > eps {
			return false
		}
	}

	return true
}

// Matrix4PolarFactor computes the Q factor in M = QS where S is a symmetric matrix and Q is
// a pure rotation. S can then be calculated from S = Q^TM.
// Returns Q,false if unable to decompose (either singular matrix detected or taking
// too long to converge). det(m) should be +ve. If return is false then Q is undefined.
func Matrix4PolarFactor(m Matrix4) (Matrix4, bool) {
	Q := m

	for i := 0; i < 10; i++ {
		Qinv, ok := Matrix4Inverse(Q)

		if !ok {
			return Qinv, false
		}

		Qnew := Matrix4Scale(.5, Matrix4Add(Q, Matrix4Transpose(Qinv)))

		if Matrix4Eq(Qnew, Q, 0.000001) {
			return Qnew, true
		}

		Q = Qnew
	}

	return Q, false
}
