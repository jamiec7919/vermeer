// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"fmt"
)

// Quat represents a quaternion.
type Quat struct {
	X, Y, Z, W float32
}

// QuatIdentity is the identity quaternion.
func QuatIdentity() Quat { return Quat{0, 0, 0, 1} }

// String returns a string representation of the quaternion: [x,y,z,w].
func (q *Quat) String() string {
	return fmt.Sprintf("[%v,%v,%v,%v]", q.X, q.Y, q.Z, q.W)
}

// QuatLength returns the length of the quaternion.
func QuatLength(q Quat) float32 {
	return Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)
}

// QuatMul returns the product of two quaternions.
func QuatMul(a, b Quat) Quat {
	var o Quat

	o.W = a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z
	o.X = a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y
	o.Y = a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X
	o.Z = a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W
	return o
}

// QuatAdd returns the sum of two quaternions.
func QuatAdd(a, b Quat) Quat {
	var o Quat

	o.W = a.W + b.W
	o.X = a.X + b.X
	o.Y = a.Y + b.Y
	o.Z = a.Z + b.Z
	return o
}

// QuatScale scales the given quaternion by a scalar.
func QuatScale(s float32, q Quat) Quat {
	var o Quat

	o.X = q.X * s
	o.Y = q.Y * s
	o.Z = q.Z * s
	o.W = q.W * s
	return o
}

// QuatSlerp performs a sperical linear inerpolation between qa and qb parameterised by t.
func QuatSlerp(qa, qb Quat, t float32) Quat {
	var qm Quat

	// Calculate angle between them.
	cosHalfTheta := qa.W*qb.W + qa.X*qb.X + qa.Y*qb.Y + qa.Z*qb.Z

	// if qa=qb or qa=-qb then theta = 0 and we can return qa
	if Abs(cosHalfTheta) >= 1.0 {
		qm.W = qa.W
		qm.X = qa.X
		qm.Y = qa.Y
		qm.Z = qa.Z
		return qm
	}

	// Calculate temporary values.
	halfTheta := Acos(cosHalfTheta)
	sinHalfTheta := Sqrt(1.0 - cosHalfTheta*cosHalfTheta)

	// if theta = 180 degrees then result is not fully defined
	// we could rotate around any axis normal to qa or qb
	if Abs(sinHalfTheta) < 0.001 { // fabs is floating point absolute
		qm.W = (qa.W*0.5 + qb.W*0.5)
		qm.X = (qa.X*0.5 + qb.X*0.5)
		qm.Y = (qa.Y*0.5 + qb.Y*0.5)
		qm.Z = (qa.Z*0.5 + qb.Z*0.5)
		return qm
	}

	ratioA := Sin((1-t)*halfTheta) / sinHalfTheta
	ratioB := Sin(t*halfTheta) / sinHalfTheta

	//calculate Quaternion.
	qm.W = (qa.W*ratioA + qb.W*ratioB)
	qm.X = (qa.X*ratioA + qb.X*ratioB)
	qm.Y = (qa.Y*ratioA + qb.Y*ratioB)
	qm.Z = (qa.Z*ratioA + qb.Z*ratioB)
	return qm
}

// Normalize in-place normalizes the quaternion.
func (q *Quat) Normalize() {
	n := 1.0 / Sqrt(q.X*q.X+q.Y*q.Y+q.Z*q.Z+q.W*q.W)

	q.X = q.X * n
	q.Y = q.Y * n
	q.Z = q.Z * n
	q.W = q.W * n

}

// QuatToMatrix4 returns the matrix representation of the given quaternion.
func QuatToMatrix4(q Quat) Matrix4 {
	var m Matrix4

	n := 1.0 / Sqrt(q.X*q.X+q.Y*q.Y+q.Z*q.Z+q.W*q.W)

	qX := q.X * n
	qY := q.Y * n
	qZ := q.Z * n
	qW := q.W * n

	m.Set(0, 0, 1-2*qY*qY-2*qZ*qZ)
	m.Set(0, 1, 2*qX*qY-2*qW*qZ)
	m.Set(0, 2, 2*qX*qZ+2*qW*qY)

	m.Set(1, 0, 2*qX*qY+2*qW*qZ)
	m.Set(1, 1, 1-2*qX*qX-2*qZ*qZ)
	m.Set(1, 2, 2*qY*qZ-2*qW*qX)

	m.Set(2, 0, 2*qX*qZ-2*qW*qY)
	m.Set(2, 1, 2*qY*qZ+2*qW*qX)
	m.Set(2, 2, 1-2*qX*qX-2*qY*qY)

	m.Set(3, 3, 1.0)

	return m
}

// Matrix4ToQuat converts the upper left 3x3 part of the 4x4 matrix m into a
// quaternion.  Assumes m is an orthonormal rotation matrix (doesn't have to
// be quite there, but normalize quaternion afterwards).
func Matrix4ToQuat(m Matrix4) Quat {
	var q Quat

	tr := m.Elt(0, 0) + m.Elt(1, 1) + m.Elt(2, 2)

	if tr > 0.0 {
		S := Sqrt(tr+1.0) * 2 // S=4*qw
		q.W = 0.25 * S
		q.X = (m.Elt(2, 1) - m.Elt(1, 2)) / S
		q.Y = (m.Elt(0, 2) - m.Elt(2, 0)) / S
		q.Z = (m.Elt(1, 0) - m.Elt(0, 1)) / S
	} else if (m.Elt(0, 0) > m.Elt(1, 1)) && (m.Elt(0, 0) > m.Elt(2, 2)) {
		S := Sqrt(1.0+m.Elt(0, 0)-m.Elt(1, 1)-m.Elt(2, 2)) * 2 // S=4*qx
		q.W = (m.Elt(2, 1) - m.Elt(1, 2)) / S
		q.X = 0.25 * S
		q.Y = (m.Elt(0, 1) + m.Elt(1, 0)) / S
		q.Z = (m.Elt(0, 2) + m.Elt(2, 0)) / S
	} else if m.Elt(1, 1) > m.Elt(2, 2) {
		S := Sqrt(1.0+m.Elt(1, 1)-m.Elt(0, 0)-m.Elt(2, 2)) * 2 // S=4*qy
		q.W = (m.Elt(0, 2) - m.Elt(2, 0)) / S
		q.X = (m.Elt(0, 1) + m.Elt(1, 0)) / S
		q.Y = 0.25 * S
		q.Z = (m.Elt(1, 2) + m.Elt(2, 1)) / S
	} else {
		S := Sqrt(1.0+m.Elt(2, 2)-m.Elt(0, 0)-m.Elt(1, 1)) * 2 // S=4*qz
		q.W = (m.Elt(1, 0) - m.Elt(0, 1)) / S
		q.X = (m.Elt(0, 2) + m.Elt(2, 0)) / S
		q.Y = (m.Elt(1, 2) + m.Elt(2, 1)) / S
		q.Z = 0.25 * S
	}
	return q
}
