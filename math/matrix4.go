// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/*
	Column major

	[ 0 4 8 12 ]
	[ 1 5 9 13 ]
	[ 2 6 10 14]
	[ 3 7 11 15]

	Transforms POST multiply matrix by column vector.

*/

type Matrix4 [16]float32

func (m *Matrix4) Elt(i, j int) float32 {
	return m[(j*4)+i]
}

var Matrix4Identity = Matrix4{1.0, 0.0, 0.0, 0.0,
	0.0, 1.0, 0.0, 0.0,
	0.0, 0.0, 1.0, 0.0,
	0.0, 0.0, 0.0, 1.0}

var Matrix4Null = Matrix4{0.0, 0.0, 0.0, 0.0,
	0.0, 0.0, 0.0, 0.0,
	0.0, 0.0, 0.0, 0.0,
	0.0, 0.0, 0.0, 0.0}

func Matrix4Add(a, b Matrix4) (c Matrix4) {
	for i := range c {
		c[i] = a[i] + b[i]
	}

	return
}

func Matrix4Sub(a, b Matrix4) (c Matrix4) {
	for i := range c {
		c[i] = a[i] - b[i]
	}

	return
}

func Matrix4Mul(a, b Matrix4) (c Matrix4) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				c[(j*4)+i] += a[(k*4)+i] * b[(j*4)+k]
			}
		}
	}

	return
}

func Matrix4Transpose(a Matrix4) (c Matrix4) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			c[(i*4)+j] = a[(j*4)+i]
		}
	}

	return
}

func Matrix4Inverse(m Matrix4) (c Matrix4, ok bool) {
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
		return c, false
	}

	det = 1.0 / det

	for i := 0; i < 16; i++ {
		c[i] = inv[i] * det
	}

	return c, true
}

func Matrix4MulPoint(a Matrix4, b Vec3) (c Vec3) {
	c[0] = a[0]*b[0] + a[4]*b[1] + a[8]*b[2] + a[12]
	c[1] = a[1]*b[0] + a[5]*b[1] + a[9]*b[2] + a[13]
	c[2] = a[2]*b[0] + a[6]*b[1] + a[10]*b[2] + a[14]

	return
}

func _Matrix4MulPoint2(a Matrix4, b Vec3) (c Vec3) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			c[i] += a[(j*4)+i] * b[j]
		}
		c[i] += a[(3*4)+i] // Last element
	}

	return
}

func Matrix4MulHPoint(a Matrix4, b [4]float32) (c [4]float32) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			c[i] += a[(j*4)+i] * b[j]
		}
	}

	return
}

func Matrix4MulVec(a Matrix4, b Vec3) (c Vec3) {
	c[0] = a[0]*b[0] + a[4]*b[1] + a[8]*b[2]
	c[1] = a[1]*b[0] + a[5]*b[1] + a[9]*b[2]
	c[2] = a[2]*b[0] + a[6]*b[1] + a[10]*b[2]
	return
}

func _Matrix4MulVec2(a Matrix4, b Vec3) (c Vec3) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			c[i] += a[(j*4)+i] * b[j]
		}
	}

	return
}

func Matrix4Translate(X, Y, Z float32) (c Matrix4) {
	c = Matrix4Identity
	c[12] = -X
	c[13] = -Y
	c[14] = -Z
	return
}

func Matrix4Scale(X, Y, Z float32) (c Matrix4) {
	c = Matrix4Identity
	c[0] = X
	c[5] = Y
	c[10] = Z
	return
}

func Matrix4Rotate(angle, X, Y, Z float32) (c Matrix4) {
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

	return
}
