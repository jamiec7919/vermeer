// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mesh

import (
	//	"log"
	//"unsafe"
	//"fmt"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

const CHECK_EMPTY_LEAF = true

const VisRayEpsilon float32 = 0.0001

const BACKFACE_CULL = false

//go:nosplit
//go:noescape
func rayNodeIntersectAll_asm(ray *core.Ray, node *qbvh.Node, hit *[4]int32, tNear *[4]float32)

//go:nosplit
func rayNodeIntersectAll(ray *core.Ray, node *qbvh.Node, hit *[4]int32, tNear *[4]float32) {
	//idx+(i*12)+(k*4) = bounds[i][k]
	for k := 0; k < 4; k++ {
		t1 := (node.Boxes[k+int(0*12)+(0*4)] - ray.P[0]) * ray.Dinv[0] // idx+0
		t2 := (node.Boxes[k+int(1*12)+(0*4)] - ray.P[0]) * ray.Dinv[0] // idx+12

		tmin := m.Min(t1, t2)
		tmax := m.Max(t1, t2)

		t1 = (node.Boxes[k+int(0*12)+(1*4)] - ray.P[1]) * ray.Dinv[1] // idx+4
		t2 = (node.Boxes[k+int(1*12)+(1*4)] - ray.P[1]) * ray.Dinv[1] // idx+16

		//tmin = m.Max(tmin, m.Min(m.Min(t1, t2), tmax))
		//tmax = m.Min(tmax, m.Max(m.Max(t1, t2), tmin))
		tmin = m.Max(tmin, m.Min(t1, t2))
		tmax = m.Min(tmax, m.Max(t1, t2))

		t1 = (node.Boxes[k+int(0*12)+(2*4)] - ray.P[2]) * ray.Dinv[2] // idx+8
		t2 = (node.Boxes[k+int(1*12)+(2*4)] - ray.P[2]) * ray.Dinv[2] // idx+20

		//tmin = m.Max(tmin, m.Min(m.Min(t1, t2), tmax))
		//tmax = m.Min(tmax, m.Max(m.Max(t1, t2), tmin))
		tmin = m.Max(tmin, m.Min(t1, t2))
		tmax = m.Min(tmax, m.Max(t1, t2))

		(*tNear)[k] = m.Max(tmin, 0)
		//l og.Printf("%v %v", tNear, tFar)
		if tmax >= tNear[k] {
			(*hit)[k] = 1
		} else {
			(*hit)[k] = 0
		}
	}
}

//go:nosplit
//go:noescape
// This is approx 2x the speeeed of the Go version
func rayNodeIntersect2(ray *core.Ray, node *qbvh.Node, idx int) (bool, float32)

//go:nosplit
func rayNodeIntersect(ray *core.Ray, node *qbvh.Node, idx int) (bool, float32) {
	//idx+(i*12)+(k*4) = bounds[i][k]

	t1 := (node.Boxes[idx+int(0*12)+(0*4)] - ray.P[0]) * ray.Dinv[0] // idx+0
	t2 := (node.Boxes[idx+int(1*12)+(0*4)] - ray.P[0]) * ray.Dinv[0] // idx+12

	tmin := m.Min(t1, t2)
	tmax := m.Max(t1, t2)

	t1 = (node.Boxes[idx+int(0*12)+(1*4)] - ray.P[1]) * ray.Dinv[1] // idx+4
	t2 = (node.Boxes[idx+int(1*12)+(1*4)] - ray.P[1]) * ray.Dinv[1] // idx+16

	//tmin = m.Max(tmin, m.Min(m.Min(t1, t2), tmax))
	//tmax = m.Min(tmax, m.Max(m.Max(t1, t2), tmin))
	tmin = m.Max(tmin, m.Min(t1, t2))
	tmax = m.Min(tmax, m.Max(t1, t2))

	t1 = (node.Boxes[idx+int(0*12)+(2*4)] - ray.P[2]) * ray.Dinv[2] // idx+8
	t2 = (node.Boxes[idx+int(1*12)+(2*4)] - ray.P[2]) * ray.Dinv[2] // idx+20

	//tmin = m.Max(tmin, m.Min(m.Min(t1, t2), tmax))
	//tmax = m.Min(tmax, m.Max(m.Max(t1, t2), tmin))
	tmin = m.Max(tmin, m.Min(t1, t2))
	tmax = m.Min(tmax, m.Max(t1, t2))

	tNear := m.Max(tmin, 0)
	//l og.Printf("%v %v", tNear, tFar)
	return tmax >= tNear, tNear
}

// Some meshes just seem to need an epsilon to work (e.g. hairball)
//go:nosplit
func visIntersectFaceEpsilon(ray *core.RayData, face *FaceGeom, epsilon float32) bool {
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
	/*
	 */
	// else
	/* if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
		return
	}*/
	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.Ray.S[2] * A_Kz
	//Bz := ray.Ray.S[2] * B_Kz
	//Cz := ray.Ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)
	//T := U*Az + V*Bz + W*Cz

	// Backface cull:
	if BACKFACE_CULL {
		if T < 0.0 || T > ray.Ray.Tclosest*det {
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

//go:nosplit
func visIntersectFace(ray *core.RayData, face *FaceGeom) bool {
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
	/*
	 */
	// else
	/* if (U < 0.0 || V < 0.0 || W < 0.0) && (U > 0.0 || V > 0.0 || W > 0.0) {
		return
	}*/
	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return false
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.Ray.S[2] * A_Kz
	//Bz := ray.Ray.S[2] * B_Kz
	//Cz := ray.Ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)
	//T := U*Az + V*Bz + W*Cz

	// Backface cull:
	if BACKFACE_CULL {
		if T < 0.0 || T > ray.Ray.Tclosest*det {
			return false
		}
	} else {
		det_sign := m.SignMask(det)

		if m.Xorf(T, det_sign) < 0 || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
			return false
		}
	}

	return true
}

func (mesh *Mesh) visIntersectTris(ray *core.RayData, base, count int) bool {
	for i := base; i < base+count; i++ {
		if visIntersectFace(ray, &mesh.Faces[mesh.faceindex[i]]) {
			return true
		}
	}
	return false
}

//go:noslit
func traceFaceEpsilon(mesh *Mesh, ray *core.RayData, face *FaceGeom, epsilon float32) {
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
		return
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.Ray.S[2] * A_Kz
	//Bz := ray.Ray.S[2] * B_Kz
	//Cz := ray.Ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	det_sign := m.SignMask(det)

	if m.Xorf(T, det_sign) < epsilon*m.Xorf(det, det_sign) || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
		return
	}

	rcpDet := 1.0 / det

	U = U * rcpDet
	V = V * rcpDet
	W = W * rcpDet
	ray.Ray.Tclosest = T * rcpDet

	xAbsSum := m.Abs(U*face.V[0][0]) + m.Abs(V*face.V[1][0]) + m.Abs(W*face.V[2][0])
	yAbsSum := m.Abs(U*face.V[0][1]) + m.Abs(V*face.V[1][1]) + m.Abs(W*face.V[2][1])
	zAbsSum := m.Abs(U*face.V[0][2]) + m.Abs(V*face.V[1][2]) + m.Abs(W*face.V[2][2])

	// These try to fix the 'flat surface at origin' problem, if it is indeed a problem and not
	// just something stupid I'm doing.
	if xAbsSum == 0.0 {
		xAbsSum = 0.08
	}
	if yAbsSum == 0.0 {
		yAbsSum = 0.08 // empirically discovered constant that seems to work
	}
	if zAbsSum == 0.0 {
		zAbsSum = 0.08
	}

	pError := m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})

	nAbs := m.Vec3{m.Abs(face.N[0]), m.Abs(face.N[1]), m.Abs(face.N[2])}
	d := m.Vec3Dot(nAbs, pError)

	offset := m.Vec3Scale(d, face.N)

	if m.Vec3Dot(ray.Ray.D, face.N) > 0 { // Is it a back face hit?
		offset = m.Vec3Neg(offset)
	}
	p := m.Vec3Add3(m.Vec3Scale(U, face.V[0]), m.Vec3Scale(V, face.V[1]), m.Vec3Scale(W, face.V[2]))
	/*
		po := m.Vec3Add(p, offset)

		// round po away from p
		for i := range po {
			//log.Printf("%v %v %v", i, offset[i], po[i])
			if offset[i] > 0 {
				po[i] = m.NextFloatUp(po[i])
			} else if offset[i] < 0 {
				po[i] = m.NextFloatDown(po[i])
			}
			//log.Printf("%v %v", i, po[i])
		}*/
	ray.Result.POffset = offset
	ray.Result.P = p

	ray.Result.Ng = face.N
	//ray.Result.Tg = m.Vec3Normalize(m.Vec3Cross(face.N, m.Vec3Normalize(m.Vec3Sub(face.V[2], face.V[0]))))
	//ray.Result.Bg = m.Vec3Cross(face.N, ray.Result.Tg)

	if mesh.Vn != nil {
		ray.Result.Ns = m.Vec3Add3(m.Vec3Scale(U, mesh.Vn[face.Vi[0]]), m.Vec3Scale(V, mesh.Vn[face.Vi[1]]), m.Vec3Scale(W, mesh.Vn[face.Vi[2]]))
	} else {
		ray.Result.Ns = ray.Result.Ng
	}

	for k := range mesh.Vuv {
		if k >= len(ray.Result.UV) { // Would need to allocate.., could swap to a different allocated set if this occurs
			// panic("ray->tri intersect: not implemented UV count > " + string(len(ray.Result.UV)))
			break
		}

		if mesh.Vuv[k] != nil {
			ray.Result.UV[k][0] = U*mesh.Vuv[k][face.Vi[0]][0] + V*mesh.Vuv[k][face.Vi[1]][0] + W*mesh.Vuv[k][face.Vi[2]][0]
			ray.Result.UV[k][1] = U*mesh.Vuv[k][face.Vi[0]][1] + V*mesh.Vuv[k][face.Vi[1]][1] + W*mesh.Vuv[k][face.Vi[2]][1]

			s1 := mesh.Vuv[k][face.Vi[1]][0] - mesh.Vuv[k][face.Vi[0]][0]
			t1 := mesh.Vuv[k][face.Vi[1]][1] - mesh.Vuv[k][face.Vi[0]][1]
			s2 := mesh.Vuv[k][face.Vi[2]][0] - mesh.Vuv[k][face.Vi[0]][0]
			t2 := mesh.Vuv[k][face.Vi[2]][1] - mesh.Vuv[k][face.Vi[0]][1]

			det := 1.0 / (s1*t2 - s2*t1)
			ray.Result.Pu[k][0] = det * (t2*(face.V[1][0]-face.V[0][0]) - t1*(face.V[2][0]-face.V[0][0]))
			ray.Result.Pu[k][1] = det * (t2*(face.V[1][1]-face.V[0][1]) - t1*(face.V[2][1]-face.V[0][1]))
			ray.Result.Pu[k][2] = det * (t2*(face.V[1][2]-face.V[0][2]) - t1*(face.V[2][2]-face.V[0][2]))
			ray.Result.Pv[k][0] = det * (-s2*(face.V[1][0]-face.V[0][0]) + s1*(face.V[2][0]-face.V[0][0]))
			ray.Result.Pv[k][1] = det * (-s2*(face.V[1][1]-face.V[0][1]) + s1*(face.V[2][1]-face.V[0][1]))
			ray.Result.Pv[k][2] = det * (-s2*(face.V[1][2]-face.V[0][2]) + s1*(face.V[2][2]-face.V[0][2]))
		}
	}

	if mesh.Vuv == nil {
		ray.Result.UV[0][0] = U*0 + V*1 + W*0
		ray.Result.UV[0][1] = U*0 + V*0 + W*1

		s1 := float32(1 - 0)
		t1 := float32(0 - 0)
		s2 := float32(0 - 0)
		t2 := float32(1 - 0)

		det := 1.0 / (s1*t2 - s2*t1)
		ray.Result.Pu[0][0] = det * (t2*(face.V[1][0]-face.V[0][0]) - t1*(face.V[2][0]-face.V[0][0]))
		ray.Result.Pu[0][1] = det * (t2*(face.V[1][1]-face.V[0][1]) - t1*(face.V[2][1]-face.V[0][1]))
		ray.Result.Pu[0][2] = det * (t2*(face.V[1][2]-face.V[0][2]) - t1*(face.V[2][2]-face.V[0][2]))
		ray.Result.Pv[0][0] = det * (-s2*(face.V[1][0]-face.V[0][0]) + s1*(face.V[2][0]-face.V[0][0]))
		ray.Result.Pv[0][1] = det * (-s2*(face.V[1][1]-face.V[0][1]) + s1*(face.V[2][1]-face.V[0][1]))
		ray.Result.Pv[0][2] = det * (-s2*(face.V[1][2]-face.V[0][2]) + s1*(face.V[2][2]-face.V[0][2]))

	}

	ray.Result.MtlId = face.MtlId

}

//go:noslit
func traceFace(mesh *Mesh, ray *core.RayData, face *FaceGeom) {
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
		return
	}

	// Calculate determinant
	det := U + V + W

	if det == 0.0 {
		return
	}

	// Calc scaled z-coords of verts and calc the hit dis
	//Az := ray.Ray.S[2] * A_Kz
	//Bz := ray.Ray.S[2] * B_Kz
	//Cz := ray.Ray.S[2] * C_Kz

	T := ray.Ray.S[2] * (U*A_Kz + V*B_Kz + W*C_Kz)

	det_sign := m.SignMask(det)

	if m.Xorf(T, det_sign) < 0.0 || m.Xorf(T, det_sign) > ray.Ray.Tclosest*m.Xorf(det, det_sign) {
		return
	}

	rcpDet := 1.0 / det

	U = U * rcpDet
	V = V * rcpDet
	W = W * rcpDet
	ray.Ray.Tclosest = T * rcpDet

	xAbsSum := m.Abs(U*face.V[0][0]) + m.Abs(V*face.V[1][0]) + m.Abs(W*face.V[2][0])
	yAbsSum := m.Abs(U*face.V[0][1]) + m.Abs(V*face.V[1][1]) + m.Abs(W*face.V[2][1])
	zAbsSum := m.Abs(U*face.V[0][2]) + m.Abs(V*face.V[1][2]) + m.Abs(W*face.V[2][2])

	// These try to fix the 'flat surface at origin' problem, if it is indeed a problem and not
	// just something stupid I'm doing.
	if xAbsSum == 0.0 {
		xAbsSum = 0.08
	}
	if yAbsSum == 0.0 {
		yAbsSum = 0.08 // empirically discovered constant that seems to work
	}
	if zAbsSum == 0.0 {
		zAbsSum = 0.08
	}

	pError := m.Vec3Scale(m.Gamma(7), m.Vec3{xAbsSum, yAbsSum, zAbsSum})

	nAbs := m.Vec3{m.Abs(face.N[0]), m.Abs(face.N[1]), m.Abs(face.N[2])}
	d := m.Vec3Dot(nAbs, pError)

	offset := m.Vec3Scale(d, face.N)

	if m.Vec3Dot(ray.Ray.D, face.N) > 0 { // Is it a back face hit?
		offset = m.Vec3Neg(offset)
	}
	p := m.Vec3Add3(m.Vec3Scale(U, face.V[0]), m.Vec3Scale(V, face.V[1]), m.Vec3Scale(W, face.V[2]))
	/*
		po := m.Vec3Add(p, offset)

		// round po away from p
		for i := range po {
			//log.Printf("%v %v %v", i, offset[i], po[i])
			if offset[i] > 0 {
				po[i] = m.NextFloatUp(po[i])
			} else if offset[i] < 0 {
				po[i] = m.NextFloatDown(po[i])
			}
			//log.Printf("%v %v", i, po[i])
		}*/
	ray.Result.POffset = offset
	ray.Result.P = p

	ray.Result.Ng = face.N
	//ray.Result.Tg = m.Vec3Normalize(m.Vec3Cross(face.N, m.Vec3Normalize(m.Vec3Sub(face.V[2], face.V[0]))))
	//ray.Result.Bg = m.Vec3Cross(face.N, ray.Result.Tg)

	if mesh.Vn != nil {
		ray.Result.Ns = m.Vec3Add3(m.Vec3Scale(U, mesh.Vn[face.Vi[0]]), m.Vec3Scale(V, mesh.Vn[face.Vi[1]]), m.Vec3Scale(W, mesh.Vn[face.Vi[2]]))
	} else {
		ray.Result.Ns = ray.Result.Ng
	}

	//log.Printf("%v %v", ray.Result.Ng, ray.Result.Ns)

	for k := range mesh.Vuv {
		if k >= len(ray.Result.UV) { // Would need to allocate.., could swap to a different allocated set if this occurs
			// panic("ray->tri intersect: not implemented UV count > " + string(len(ray.Result.UV)))
			break
		}

		if mesh.Vuv[k] != nil {
			ray.Result.UV[k][0] = U*mesh.Vuv[k][face.Vi[0]][0] + V*mesh.Vuv[k][face.Vi[1]][0] + W*mesh.Vuv[k][face.Vi[2]][0]
			ray.Result.UV[k][1] = U*mesh.Vuv[k][face.Vi[0]][1] + V*mesh.Vuv[k][face.Vi[1]][1] + W*mesh.Vuv[k][face.Vi[2]][1]

			s1 := mesh.Vuv[k][face.Vi[1]][0] - mesh.Vuv[k][face.Vi[0]][0]
			t1 := mesh.Vuv[k][face.Vi[1]][1] - mesh.Vuv[k][face.Vi[0]][1]
			s2 := mesh.Vuv[k][face.Vi[2]][0] - mesh.Vuv[k][face.Vi[0]][0]
			t2 := mesh.Vuv[k][face.Vi[2]][1] - mesh.Vuv[k][face.Vi[0]][1]

			invDet := (s1*t2 - s2*t1)

			if invDet == 0.0 {
				//panic("Err" + fmt.Sprintf("%v", mesh.Name))
				invDet = 1.0
				s1 = 1.0
				t1 = 0.0
				s2 = 0.0
				t2 = 1.0
			}
			det := 1.0 / invDet
			ray.Result.Pu[k][0] = det * (t2*(face.V[1][0]-face.V[0][0]) - t1*(face.V[2][0]-face.V[0][0]))
			ray.Result.Pu[k][1] = det * (t2*(face.V[1][1]-face.V[0][1]) - t1*(face.V[2][1]-face.V[0][1]))
			ray.Result.Pu[k][2] = det * (t2*(face.V[1][2]-face.V[0][2]) - t1*(face.V[2][2]-face.V[0][2]))
			ray.Result.Pv[k][0] = det * (-s2*(face.V[1][0]-face.V[0][0]) + s1*(face.V[2][0]-face.V[0][0]))
			ray.Result.Pv[k][1] = det * (-s2*(face.V[1][1]-face.V[0][1]) + s1*(face.V[2][1]-face.V[0][1]))
			ray.Result.Pv[k][2] = det * (-s2*(face.V[1][2]-face.V[0][2]) + s1*(face.V[2][2]-face.V[0][2]))
		}
	}

	if mesh.Vuv == nil {
		ray.Result.UV[0][0] = U*0 + V*1 + W*0
		ray.Result.UV[0][1] = U*0 + V*0 + W*1

		s1 := float32(1 - 0)
		t1 := float32(0 - 0)
		s2 := float32(0 - 0)
		t2 := float32(1 - 0)

		det := 1.0 / (s1*t2 - s2*t1)
		ray.Result.Pu[0][0] = det * (t2*(face.V[1][0]-face.V[0][0]) - t1*(face.V[2][0]-face.V[0][0]))
		ray.Result.Pu[0][1] = det * (t2*(face.V[1][1]-face.V[0][1]) - t1*(face.V[2][1]-face.V[0][1]))
		ray.Result.Pu[0][2] = det * (t2*(face.V[1][2]-face.V[0][2]) - t1*(face.V[2][2]-face.V[0][2]))
		ray.Result.Pv[0][0] = det * (-s2*(face.V[1][0]-face.V[0][0]) + s1*(face.V[2][0]-face.V[0][0]))
		ray.Result.Pv[0][1] = det * (-s2*(face.V[1][1]-face.V[0][1]) + s1*(face.V[2][1]-face.V[0][1]))
		ray.Result.Pv[0][2] = det * (-s2*(face.V[1][2]-face.V[0][2]) + s1*(face.V[2][2]-face.V[0][2]))

	}

	ray.Result.MtlId = face.MtlId

}

//go:nosplit
func (mesh *Mesh) traceRayAccelIndexed(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

			if m.SignMask(ray.Ray.D[pnode.Axis1]) != (1 << 31) {
				order[0], order[1] = order[1], order[0]
			}

			if m.SignMask(ray.Ray.D[pnode.Axis2]) != (1 << 31) {
				order[2], order[3] = order[3], order[2]
			}

			if m.SignMask(ray.Ray.D[pnode.Axis0]) != (1 << 31) {
				order[0], order[2] = order[2], order[0]
				order[1], order[3] = order[3], order[1]
			}

			for j := range order {
				k := order[j]

				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				} else {
					//log.Printf("Miss %v %v", node, pnode.Children[k])
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				face := &mesh.Faces[mesh.faceindex[i]]

				traceFace(mesh, ray, face)
			}
		}
	}

}

//go:nosplit
func (mesh *Mesh) traceRayAccelIndexedEpsilon(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

			if ray.Ray.D[pnode.Axis0] < 0 {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[3] = 3
					order[2] = 2
				} else {
					order[3] = 2
					order[2] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[1] = 1
					order[0] = 0
				} else {
					order[1] = 0
					order[0] = 1
				}
			} else {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[1] = 3
					order[0] = 2
				} else {
					order[1] = 2
					order[0] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[3] = 1
					order[2] = 0
				} else {
					order[3] = 0
					order[2] = 1
				}

			}

			for j := range order {
				k := order[j]
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				} else {
					//log.Printf("Miss %v %v", node, pnode.Children[k])
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				face := &mesh.Faces[mesh.faceindex[i]]

				traceFaceEpsilon(mesh, ray, face, mesh.RayBias)
			}

		}
	}

}

//go:nosplit
func (mesh *Mesh) visRayAccelIndexed(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				if visIntersectFace(ray, &mesh.Faces[mesh.faceindex[i]]) {
					ray.Ray.Tclosest = 0.5
					return
				}
			}
		}
	}

}

//go:nosplit
func (mesh *Mesh) visRayAccelIndexedEpsilon(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				if visIntersectFaceEpsilon(ray, &mesh.Faces[mesh.faceindex[i]], mesh.RayBias) {
					ray.Ray.Tclosest = 0.5
					return
				}
			}
		}
	}

}

//go:nosplit
func (mesh *Mesh) traceRayAccel(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

			if m.SignMask(ray.Ray.D[pnode.Axis1]) != (1 << 31) {
				order[0], order[1] = order[1], order[0]
			}

			if m.SignMask(ray.Ray.D[pnode.Axis2]) != (1 << 31) {
				order[2], order[3] = order[3], order[2]
			}

			if m.SignMask(ray.Ray.D[pnode.Axis0]) != (1 << 31) {
				order[0], order[2] = order[2], order[0]
				order[1], order[3] = order[3], order[1]
			}

			for j := range order {
				k := order[j]

				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				} else {
					//log.Printf("Miss %v %v", node, pnode.Children[k])
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				face := &mesh.Faces[i]

				traceFace(mesh, ray, face)
			}
		}
	}

}

//go:nosplit
func (mesh *Mesh) traceRayAccelEpsilon(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

			if ray.Ray.D[pnode.Axis0] < 0 {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[3] = 3
					order[2] = 2
				} else {
					order[3] = 2
					order[2] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[1] = 1
					order[0] = 0
				} else {
					order[1] = 0
					order[0] = 1
				}
			} else {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[1] = 3
					order[0] = 2
				} else {
					order[1] = 2
					order[0] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[3] = 1
					order[2] = 0
				} else {
					order[3] = 0
					order[2] = 1
				}

			}

			for j := range order {
				k := order[j]
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				} else {
					//log.Printf("Miss %v %v", node, pnode.Children[k])
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				face := &mesh.Faces[i]

				traceFaceEpsilon(mesh, ray, face, mesh.RayBias)
			}

		}
	}

}

//go:nosplit
func (mesh *Mesh) visRayAccel(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				if visIntersectFace(ray, &mesh.Faces[i]) {
					ray.Ray.Tclosest = 0.5
					return
				}
			}
		}
	}

}

//go:nosplit
func (mesh *Mesh) visRayAccelEpsilon(ray *core.RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				if visIntersectFaceEpsilon(ray, &mesh.Faces[i], mesh.RayBias) {
					ray.Ray.Tclosest = 0.5
					return
				}
			}
		}
	}

}
