// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

/* Matrix3 represents a 3x3 matrix. FIXME: THIS IS NOT TRUE.

Column major

[ 0 4 8 ]
[ 1 5 9 ]
[ 2 6 10 ]

Transforms must POST multiply matrix by column vector.
*/
type Matrix3 [9]float32

// Matrix3Transpose returns the transpose of 4x4 matrix a.
func Matrix3Transpose(a Matrix3) Matrix3 {
	var c Matrix3

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			c[(i*3)+j] = a[(j*3)+i]
		}
	}

	return c
}

// Matrix3Inverse returns the matrix inverse of 4x4 matrix a.
func Matrix3Inverse(m Matrix3) (Matrix3, bool) {
	// computes the inverse of a matrix m
	det := m[0]*(m[4]*m[8]-m[7]*m[5]) -
		m[1]*(m[3]*m[8]-m[5]*m[6]) +
		m[2]*(m[3]*m[7]-m[4]*m[6])

	invdet := 1 / det

	var inv Matrix3

	inv[0] = (m[4]*m[8] - m[7]*m[5]) * invdet
	inv[1] = (m[2]*m[7] - m[1]*m[8]) * invdet
	inv[2] = (m[1]*m[5] - m[2]*m[4]) * invdet
	inv[3] = (m[5]*m[6] - m[3]*m[8]) * invdet
	inv[4] = (m[0]*m[8] - m[2]*m[6]) * invdet
	inv[5] = (m[3]*m[2] - m[0]*m[5]) * invdet
	inv[6] = (m[3]*m[7] - m[6]*m[4]) * invdet
	inv[7] = (m[6]*m[1] - m[0]*m[7]) * invdet
	inv[8] = (m[0]*m[4] - m[3]*m[1]) * invdet
	return inv, true
}

// Matrix3Det returns the determinant of 4x4 matrix a.
func Matrix3Det(m Matrix4) float32 {
	return 0
}
