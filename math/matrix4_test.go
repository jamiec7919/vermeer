// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "testing"

func TestMatrix4Quat(t *testing.T) {
	A := Matrix4Rotate(0.3, 1, 0, 0)
	B := Matrix4Rotate(0.5, 0, 0, 1)

	QA := Matrix4ToQuat(A)
	QB := Matrix4ToQuat(B)
	t.Logf("A: %v QA: %v", A, QA)
	t.Logf("B: %v QB: %v", B, QB)
	t.Logf("%v", QuatToMatrix4(QA))
	t.Logf("%v", QuatToMatrix4(QB))
}

func TestMatrix4PolarFactor(t *testing.T) {

	A := Matrix4Rotate(0.3, 1, 0, 0)

	t.Logf("A: %v", A)

	B := Matrix4TransformScale(.5, .1, 4)

	t.Logf("B: %v", B)

	C := Matrix4Mul(A, B)

	t.Logf("C = AB: %v", C)

	Q, ok := Matrix4PolarFactor(C)

	t.Logf("Q: %v %v", ok, Q)

	S := Matrix4Mul(Matrix4Transpose(Q), C)

	t.Logf("S = Q^T C: %v", S)

	M := Matrix4Mul(Q, S)

	t.Logf("M = QS: %v", M)

	t.Logf("M == C? %v", Matrix4Eq(C, M, 0.0001))
}
