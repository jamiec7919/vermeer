// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

type Face struct {
	V      [3]m.Vec3
	Ns     [3]m.Vec3
	UV     [3]m.Vec2
	N      m.Vec3
	PrimId uint64
}

func (f *Face) setup() {
	f.N = m.Vec3Normalize(m.Vec3Cross(m.Vec3Sub(f.V[1], f.V[0]), m.Vec3Sub(f.V[2], f.V[0])))

	for i := range f.N {
		if f.N[i] == 0.0 {
			f.N[i] = 0
		}
	}
}

const BACKFACE_CULL bool = false

func (face *Face) Bounds() (box m.BoundingBox) {
	box.Reset()

	for k := range face.V {
		box.GrowVec3(face.V[k])
	}

	return
}

func (face *Face) Centroid() (c m.Vec3) {
	for i := range c {
		c[i] = (face.V[0][i] + face.V[1][i] + face.V[2][i]) / 3.0
	}

	return
}

func (face *Face) shaderParams(ray *core.RayData, sg *core.ShaderGlobals) {
	W := 1.0 - ray.Result.Bu - ray.Result.Bv

	sg.U = ray.Result.Bu*face.UV[0][0] + ray.Result.Bv*face.UV[1][0] + W*face.UV[2][0]
	sg.V = ray.Result.Bu*face.UV[0][1] + ray.Result.Bv*face.UV[1][1] + W*face.UV[2][1]

	s1 := face.UV[1][0] - face.UV[0][0]
	t1 := face.UV[1][1] - face.UV[0][1]
	s2 := face.UV[2][0] - face.UV[0][0]
	t2 := face.UV[2][1] - face.UV[0][1]

	det := 1.0 / (s1*t2 - s2*t1)
	sg.DdPdu[0] = det * (t2*(face.V[1][0]-face.V[0][0]) - t1*(face.V[2][0]-face.V[0][0]))
	sg.DdPdu[1] = det * (t2*(face.V[1][1]-face.V[0][1]) - t1*(face.V[2][1]-face.V[0][1]))
	sg.DdPdu[2] = det * (t2*(face.V[1][2]-face.V[0][2]) - t1*(face.V[2][2]-face.V[0][2]))
	sg.DdPdu[0] = det * (-s2*(face.V[1][0]-face.V[0][0]) + s1*(face.V[2][0]-face.V[0][0]))
	sg.DdPdu[1] = det * (-s2*(face.V[1][1]-face.V[0][1]) + s1*(face.V[2][1]-face.V[0][1]))
	sg.DdPdu[2] = det * (-s2*(face.V[1][2]-face.V[0][2]) + s1*(face.V[2][2]-face.V[0][2]))

	ray.Result.Ns[0] = ray.Result.Bu*face.Ns[0][0] + ray.Result.Bv*face.Ns[1][0] + W*face.Ns[2][0]
	ray.Result.Ns[1] = ray.Result.Bu*face.Ns[0][1] + ray.Result.Bv*face.Ns[1][1] + W*face.Ns[2][1]
	ray.Result.Ns[2] = ray.Result.Bu*face.Ns[0][2] + ray.Result.Bv*face.Ns[1][2] + W*face.Ns[2][2]
	ray.Result.Ng = face.N
}

//go:noslit
func (face *Face) IntersectRay(ray *core.RayData) bool {
	A_Kz := face.V[0][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]
	B_Kz := face.V[1][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]
	C_Kz := face.V[2][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]

	var U, V, W float32

	{
		Cx := (face.V[2][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*C_Kz
		By := (face.V[1][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*B_Kz
		Cy := (face.V[2][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*C_Kz
		Bx := (face.V[1][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*B_Kz
		Ax := (face.V[0][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*A_Kz
		Ay := (face.V[0][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*A_Kz

		// Calc scaled barycentric
		U = Cx*By - Cy*Bx
		V = Ax*Cy - Ay*Cx
		W = Bx*Ay - By*Ax

		// Fallback to double precision if float edge tests fail
		if U == 0.0 || V == 0.0 || W == 0.0 {
			CxBy := float64(Cx) * float64(By)
			CyBx := float64(Cy) * float64(Bx)
			U = float32(CxBy - CyBx)

			AxCy := float64(Ax) * float64(Cy)
			AyCx := float64(Ay) * float64(Cx)
			V = float32(AxCy - AyCx)

			BxAy := float64(Bx) * float64(Ay)
			ByAx := float64(By) * float64(Ax)
			W = float32(BxAy - ByAx)

		}

	}

	if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
		return false
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.S[2] * A_Kz
	//Bz := ray.S[2] * B_Kz
	//Cz := ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	det_sign := m.SignMask(det)

	if m.Xorf(T, det_sign) < 0.0 || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
		return false
	}

	rcpDet := 1.0 / det

	U = U * rcpDet
	V = V * rcpDet
	W = W * rcpDet
	ray.Ray.Tclosest = T * rcpDet

	xAbsSum := m.Abs(U*face.V[0][0]) + m.Abs(V*face.V[1][0]) + m.Abs(W*face.V[2][0])
	yAbsSum := m.Abs(U*face.V[0][1]) + m.Abs(V*face.V[1][1]) + m.Abs(W*face.V[2][1])
	zAbsSum := m.Abs(U*face.V[0][2]) + m.Abs(V*face.V[1][2]) + m.Abs(W*face.V[2][2])

	xAbsSum = m.Max(xAbsSum, 0.08)
	yAbsSum = m.Max(yAbsSum, 0.08)
	zAbsSum = m.Max(zAbsSum, 0.08)

	//	pError := m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})
	face.setup()
	//nAbs := m.Vec3{m.Abs(face.N[0]), m.Abs(face.N[1]), m.Abs(face.N[2])}
	//d := m.Vec3Dot(nAbs, pError)
	//	d := pError[0]*m.Abs(face.N[0]) + pError[1]*m.Abs(face.N[1]) + pError[2]*m.Abs(face.N[2])
	d := m.Gamma(7)*xAbsSum*m.Abs(face.N[0]) + m.Gamma(7)*yAbsSum*m.Abs(face.N[1]) + m.Gamma(7)*zAbsSum*m.Abs(face.N[2])
	offset := m.Vec3Scale(d, face.N)

	if m.Vec3Dot(ray.Ray.D, face.N) > 0 { // Is it a back face hit?
		offset = m.Vec3Neg(offset)
	}

	//ray.Result.Error = m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})
	ray.Result.Ns = face.N
	ray.Result.Ng = face.N
	ray.Result.POffset = offset
	ray.Result.Bu = U
	ray.Result.Bv = V
	ray.Result.P = m.Vec3Add3(m.Vec3Scale(U, face.V[0]), m.Vec3Scale(V, face.V[1]), m.Vec3Scale(W, face.V[2]))
	//	ray.Result.UVW[0] = U
	//	ray.Result.UVW[1] = V
	//	ray.Result.UVW[2] = W

	ray.Result.ElemId = uint32(face.PrimId)
	return true
}

//go:noslit
func (face *Face) IntersectRayEpsilon(ray *core.RayData, epsilon float32) bool {
	A_Kz := face.V[0][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]
	B_Kz := face.V[1][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]
	C_Kz := face.V[2][ray.Ray.Kz] - ray.Ray.P[ray.Ray.Kz]

	var U, V, W float32

	{
		Cx := (face.V[2][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*C_Kz
		By := (face.V[1][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*B_Kz
		Cy := (face.V[2][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*C_Kz
		Bx := (face.V[1][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*B_Kz
		Ax := (face.V[0][ray.Ray.Kx] - ray.Ray.P[ray.Ray.Kx]) - ray.Ray.S[0]*A_Kz
		Ay := (face.V[0][ray.Ray.Ky] - ray.Ray.P[ray.Ray.Ky]) - ray.Ray.S[1]*A_Kz

		// Calc scaled barycentric
		U = Cx*By - Cy*Bx
		V = Ax*Cy - Ay*Cx
		W = Bx*Ay - By*Ax

		// Fallback to double precision if float edge tests fail
		if U == 0.0 || V == 0.0 || W == 0.0 {
			CxBy := float64(Cx) * float64(By)
			CyBx := float64(Cy) * float64(Bx)
			U = float32(CxBy - CyBx)

			AxCy := float64(Ax) * float64(Cy)
			AyCx := float64(Ay) * float64(Cx)
			V = float32(AxCy - AyCx)

			BxAy := float64(Bx) * float64(Ay)
			ByAx := float64(By) * float64(Ax)
			W = float32(BxAy - ByAx)

		}

	}

	if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
		return false
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.S[2] * A_Kz
	//Bz := ray.S[2] * B_Kz
	//Cz := ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	det_sign := m.SignMask(det)

	if m.Xorf(T, det_sign) < epsilon*m.Xorf(det, det_sign) || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
		return false
	}

	rcpDet := 1.0 / det

	U = U * rcpDet
	V = V * rcpDet
	W = W * rcpDet
	ray.Ray.Tclosest = T * rcpDet

	xAbsSum := m.Abs(U*face.V[0][0]) + m.Abs(V*face.V[1][0]) + m.Abs(W*face.V[2][0])
	yAbsSum := m.Abs(U*face.V[0][1]) + m.Abs(V*face.V[1][1]) + m.Abs(W*face.V[2][1])
	zAbsSum := m.Abs(U*face.V[0][2]) + m.Abs(V*face.V[1][2]) + m.Abs(W*face.V[2][2])

	xAbsSum = m.Max(xAbsSum, 0.08)
	yAbsSum = m.Max(yAbsSum, 0.08)
	zAbsSum = m.Max(zAbsSum, 0.08)

	//	pError := m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})
	face.setup()
	//nAbs := m.Vec3{m.Abs(face.N[0]), m.Abs(face.N[1]), m.Abs(face.N[2])}
	//d := m.Vec3Dot(nAbs, pError)
	//	d := pError[0]*m.Abs(face.N[0]) + pError[1]*m.Abs(face.N[1]) + pError[2]*m.Abs(face.N[2])
	d := m.Gamma(7)*xAbsSum*m.Abs(face.N[0]) + m.Gamma(7)*yAbsSum*m.Abs(face.N[1]) + m.Gamma(7)*zAbsSum*m.Abs(face.N[2])
	offset := m.Vec3Scale(d, face.N)

	if m.Vec3Dot(ray.Ray.D, face.N) > 0 { // Is it a back face hit?
		offset = m.Vec3Neg(offset)
	}

	ray.Result.Ns = face.N
	ray.Result.Ng = face.N
	//ray.Result.Error = m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})
	ray.Result.POffset = offset
	ray.Result.Bu = U
	ray.Result.Bv = V
	ray.Result.P = m.Vec3Add3(m.Vec3Scale(U, face.V[0]), m.Vec3Scale(V, face.V[1]), m.Vec3Scale(W, face.V[2]))
	//ray.Result.UVW[0] = U
	//ray.Result.UVW[1] = V
	//ray.Result.UVW[2] = W
	ray.Result.ElemId = uint32(face.PrimId)
	return true
}

// Some meshes just seem to need an epsilon to work (e.g. hairball)
//go:nosplit
func (face *Face) IntersectVisRayEpsilon(ray *core.RayData, epsilon float32) bool {
	Kz := ray.Ray.Kz
	A_Kz := face.V[0][Kz] - ray.Ray.P[Kz]
	B_Kz := face.V[1][Kz] - ray.Ray.P[Kz]
	C_Kz := face.V[2][Kz] - ray.Ray.P[Kz]

	// Ax = (V0[Kx]-O[Kx])-S[0]*(V0[Kz]-O[Kz])
	// = (V0[Kx]-O[Kx])-S[0]*V0[Kz]-S[0]*O[Kz]
	// = (V0[Kx]-S[0]*V0[Kz])-(O[Kx]+S[0]*O[Kz])
	Kx := ray.Ray.Kx
	Cx := (face.V[2][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*C_Kz
	Bx := (face.V[1][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*B_Kz
	Ax := (face.V[0][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*A_Kz

	Ky := ray.Ray.Ky
	By := (face.V[1][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*B_Kz
	Cy := (face.V[2][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*C_Kz
	Ay := (face.V[0][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*A_Kz

	// Calc scaled barycentric
	U := Cx*By - Cy*Bx
	V := Ax*Cy - Ay*Cx
	W := Bx*Ay - By*Ax

	// Fallback to double precision if float edge tests fail
	if U == 0.0 || V == 0.0 || W == 0.0 {
		CxBy := float64(Cx) * float64(By)
		CyBx := float64(Cy) * float64(Bx)
		U = float32(CxBy - CyBx)

		AxCy := float64(Ax) * float64(Cy)
		AyCx := float64(Ay) * float64(Cx)
		V = float32(AxCy - AyCx)

		BxAy := float64(Bx) * float64(Ay)
		ByAx := float64(By) * float64(Ax)
		W = float32(BxAy - ByAx)

	}

	// Perform edge tests
	// Backface cull:
	if BACKFACE_CULL {
		if U < 0.0 || V < 0.0 || W < 0.0 {
			return false
		}
	} else {
		if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
			return false
		}
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	Az := ray.Ray.S[2] * A_Kz
	Bz := ray.Ray.S[2] * B_Kz
	Cz := ray.Ray.S[2] * C_Kz

	T := U*Az + V*Bz + W*Cz
	// T := ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	// Backface cull:
	if BACKFACE_CULL {
		if T < epsilon || T > ray.Ray.Tclosest*det {
			return false
		}
	} else {
		det_sign := m.SignMask(det)

		if m.Xorf(T, det_sign) < epsilon*m.Xorf(det, det_sign) || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
			return false
		}
	}

	return true
}

func (face *Face) IntersectVisRay(ray *core.RayData) bool {
	Kz := ray.Ray.Kz
	A_Kz := face.V[0][Kz] - ray.Ray.P[Kz]
	B_Kz := face.V[1][Kz] - ray.Ray.P[Kz]
	C_Kz := face.V[2][Kz] - ray.Ray.P[Kz]

	// Ax = (V0[Kx]-O[Kx])-S[0]*(V0[Kz]-O[Kz])
	// = (V0[Kx]-O[Kx])-S[0]*V0[Kz]-S[0]*O[Kz]
	// = (V0[Kx]-S[0]*V0[Kz])-(O[Kx]+S[0]*O[Kz])
	Kx := ray.Ray.Kx
	Cx := (face.V[2][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*C_Kz
	Bx := (face.V[1][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*B_Kz
	Ax := (face.V[0][Kx] - ray.Ray.P[Kx]) - ray.Ray.S[0]*A_Kz

	Ky := ray.Ray.Ky
	By := (face.V[1][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*B_Kz
	Cy := (face.V[2][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*C_Kz
	Ay := (face.V[0][Ky] - ray.Ray.P[Ky]) - ray.Ray.S[1]*A_Kz

	// Calc scaled barycentric
	U := Cx*By - Cy*Bx
	V := Ax*Cy - Ay*Cx
	W := Bx*Ay - By*Ax

	// Fallback to double precision if float edge tests fail
	if U == 0.0 || V == 0.0 || W == 0.0 {
		CxBy := float64(Cx) * float64(By)
		CyBx := float64(Cy) * float64(Bx)
		U = float32(CxBy - CyBx)

		AxCy := float64(Ax) * float64(Cy)
		AyCx := float64(Ay) * float64(Cx)
		V = float32(AxCy - AyCx)

		BxAy := float64(Bx) * float64(Ay)
		ByAx := float64(By) * float64(Ax)
		W = float32(BxAy - ByAx)

	}

	// Perform edge tests
	// Backface cull:
	if BACKFACE_CULL {
		if U < 0.0 || V < 0.0 || W < 0.0 {
			return false
		}
	} else {
		if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
			return false
		}
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	Az := ray.Ray.S[2] * A_Kz
	Bz := ray.Ray.S[2] * B_Kz
	Cz := ray.Ray.S[2] * C_Kz

	T := U*Az + V*Bz + W*Cz
	// T := ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	// Backface cull:
	if BACKFACE_CULL {
		if T < 0.0 || T > ray.Ray.Tclosest*det {
			return false
		}
	} else {
		det_sign := m.SignMask(det)

		if m.Xorf(T, det_sign) < 0.0 || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
			return false
		}
	}

	return true

}
