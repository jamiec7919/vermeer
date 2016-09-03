// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

// Trace implements core.Primitive.
func (mesh *PolyMesh) Trace(ray *core.Ray, sg *core.ShaderContext) bool {
	if mesh.accel.qbvh != nil {
		// Static
		return qbvh.Trace(mesh.accel.qbvh, mesh, ray, sg)
	}

	// Motion
	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))

	return qbvh.TraceMotion(mesh.accel.mqbvh, time, key, key2, mesh, ray, sg)

}

// TraceElems implements qbvh.Primitive.
func (mesh *PolyMesh) TraceElems(ray *core.Ray, sg *core.ShaderContext, base, count int) bool {

	// NOTES: by unrolling many of the vector ops we avoid storing vec3's on the stack and allows it to fit
	// within the nosplit stack allowance.

	idx := int32(-1)
	var U, V, W float32

	for i := base; i < base+count; i++ {
		faceidx := mesh.accel.idx[i]

		if mesh.idxp != nil {
			i0 := int(mesh.idxp[faceidx*3+0])
			i1 := int(mesh.idxp[faceidx*3+1])
			i2 := int(mesh.idxp[faceidx*3+2])

			AKz := mesh.Verts.Elems[i0][ray.Kz] - ray.P[ray.Kz]
			BKz := mesh.Verts.Elems[i1][ray.Kz] - ray.P[ray.Kz]
			CKz := mesh.Verts.Elems[i2][ray.Kz] - ray.P[ray.Kz]

			var fU, fV, fW float32
			{
				Cx := (mesh.Verts.Elems[i2][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*CKz
				By := (mesh.Verts.Elems[i1][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*BKz
				Cy := (mesh.Verts.Elems[i2][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*CKz
				Bx := (mesh.Verts.Elems[i1][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*BKz
				Ax := (mesh.Verts.Elems[i0][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*AKz
				Ay := (mesh.Verts.Elems[i0][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*AKz

				// Calc scaled barycentric
				fU = Cx*By - Cy*Bx
				fV = Ax*Cy - Ay*Cx
				fW = Bx*Ay - By*Ax

				// Fallback to double precision if float edge tests fail
				if fU == 0.0 || fV == 0.0 || fW == 0.0 {
					CxBy := float64(Cx) * float64(By)
					CyBx := float64(Cy) * float64(Bx)
					fU = float32(CxBy - CyBx)

					AxCy := float64(Ax) * float64(Cy)
					AyCx := float64(Ay) * float64(Cx)
					fV = float32(AxCy - AyCx)

					BxAy := float64(Bx) * float64(Ay)
					ByAx := float64(By) * float64(Ax)
					fW = float32(BxAy - ByAx)

				}

			}

			if (fU < 0.0 || fV < 0.0 || fW < 0.0) && (fU > 0.0 || fV > 0.0 || fW > 0.0) {
				continue
			}

			// Calculate determinant
			det := fU + fV + fW

			if det == 0.0 {
				continue
			}

			// Calc scaled z-coords of verts and calc the hit dis
			/*
				Az := ray.S[2] * AKz
				Bz := ray.S[2] * BKz
				Cz := ray.S[2] * CKz
				T := fU*Az + fV*Bz + fW*Cz
			*/
			T := ray.S[2] * (fU*AKz + fV*BKz + fW*CKz)

			detSign := m.SignMask(det)

			// NOTE: the t < 0 bound here is a bit adhoc, it is possible to calculate tighter error bounds automatically.
			if m.Xorf(T, detSign) < mesh.RayBias*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
				continue
			}

			rcpDet := 1.0 / det

			U = fU * rcpDet
			V = fV * rcpDet
			W = fW * rcpDet

			ray.Tclosest = T * rcpDet
		} else {
			i0 := int(faceidx*3 + 0)
			i1 := int(faceidx*3 + 1)
			i2 := int(faceidx*3 + 2)

			AKz := mesh.Verts.Elems[i0][ray.Kz] - ray.P[ray.Kz]
			BKz := mesh.Verts.Elems[i1][ray.Kz] - ray.P[ray.Kz]
			CKz := mesh.Verts.Elems[i2][ray.Kz] - ray.P[ray.Kz]

			var fU, fV, fW float32
			{
				Cx := (mesh.Verts.Elems[i2][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*CKz
				By := (mesh.Verts.Elems[i1][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*BKz
				Cy := (mesh.Verts.Elems[i2][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*CKz
				Bx := (mesh.Verts.Elems[i1][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*BKz
				Ax := (mesh.Verts.Elems[i0][ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*AKz
				Ay := (mesh.Verts.Elems[i0][ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*AKz

				// Calc scaled barycentric
				fU = Cx*By - Cy*Bx
				fV = Ax*Cy - Ay*Cx
				fW = Bx*Ay - By*Ax

				// Fallback to double precision if float edge tests fail
				if fU == 0.0 || fV == 0.0 || fW == 0.0 {
					CxBy := float64(Cx) * float64(By)
					CyBx := float64(Cy) * float64(Bx)
					fU = float32(CxBy - CyBx)

					AxCy := float64(Ax) * float64(Cy)
					AyCx := float64(Ay) * float64(Cx)
					fV = float32(AxCy - AyCx)

					BxAy := float64(Bx) * float64(Ay)
					ByAx := float64(By) * float64(Ax)
					fW = float32(BxAy - ByAx)

				}

			}

			if (fU < 0.0 || fV < 0.0 || fW < 0.0) && (fU > 0.0 || fV > 0.0 || fW > 0.0) {
				continue
			}

			// Calculate determinant
			det := fU + fV + fW

			if det == 0.0 {
				continue
			}

			// Calc scaled z-coords of verts and calc the hit dis
			//Az := ray.S[2] * AKz
			//Bz := ray.S[2] * BKz
			//Cz := ray.S[2] * CKz

			T := ray.S[2] * (fU*AKz + fV*BKz + fW*CKz)

			detSign := m.SignMask(det)

			if m.Xorf(T, detSign) <= mesh.RayBias*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
				continue
			}

			rcpDet := 1.0 / det

			U = fU * rcpDet
			V = fV * rcpDet
			W = fW * rcpDet

			ray.Tclosest = T * rcpDet
		}

		// At this point we have an intersection.

		idx = faceidx
	}

	if idx == -1 {
		return false
	}

	if mesh.idxp != nil {
		i0 := int32(mesh.idxp[idx*3+0])
		i1 := int32(mesh.idxp[idx*3+1])
		i2 := int32(mesh.idxp[idx*3+2])

		xAbsSum := m.Abs(U*mesh.Verts.Elems[i0][0]) + m.Abs(V*mesh.Verts.Elems[i1][0]) + m.Abs(W*mesh.Verts.Elems[i2][0])
		yAbsSum := m.Abs(U*mesh.Verts.Elems[i0][1]) + m.Abs(V*mesh.Verts.Elems[i1][1]) + m.Abs(W*mesh.Verts.Elems[i2][1])
		zAbsSum := m.Abs(U*mesh.Verts.Elems[i0][2]) + m.Abs(V*mesh.Verts.Elems[i1][2]) + m.Abs(W*mesh.Verts.Elems[i2][2])

		//xAbsSum = m.Max(xAbsSum, 0.8)
		//yAbsSum = m.Max(yAbsSum, 0.8)
		//zAbsSum = m.Max(zAbsSum, 0.8)

		sg.Shader = mesh.shader

		e00 := mesh.Verts.Elems[i1][0] - mesh.Verts.Elems[i0][0]
		e01 := mesh.Verts.Elems[i1][1] - mesh.Verts.Elems[i0][1]
		e02 := mesh.Verts.Elems[i1][2] - mesh.Verts.Elems[i0][2]
		e10 := mesh.Verts.Elems[i2][0] - mesh.Verts.Elems[i0][0]
		e11 := mesh.Verts.Elems[i2][1] - mesh.Verts.Elems[i0][1]
		e12 := mesh.Verts.Elems[i2][2] - mesh.Verts.Elems[i0][2]

		//	e0 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+1], mesh.Verts.Elems[(idx*3)+0])
		//	e1 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+2], mesh.Verts.Elems[(idx*3)+0])
		//	N := m.Vec3Cross(e0, e1)
		var N m.Vec3

		N[0] = e01*e12 - e02*e11
		N[1] = e02*e10 - e00*e12
		N[2] = e00*e11 - e01*e10
		N = m.Vec3Normalize(N)

		if mesh.Normals.Elems != nil {
			for k := range sg.N {
				sg.N[k] = U*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][k] +
					V*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][k] +
					W*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][k]
			}

			sg.N = m.Vec3Normalize(sg.N)
		} else {
			sg.N = N
		}

		d := m.Gamma(7)*xAbsSum*m.Abs(N[0]) + m.Gamma(7)*yAbsSum*m.Abs(N[1]) + m.Gamma(7)*zAbsSum*m.Abs(N[2])
		offset := m.Vec3Scale(d, N)

		sg.Poffset = offset
		sg.Ng = N
		sg.Bu = U
		sg.Bv = V

		for k := range sg.P {
			sg.P[k] = U*mesh.Verts.Elems[i0][k] +
				V*mesh.Verts.Elems[i1][k] +
				W*mesh.Verts.Elems[i2][k]
		}
		sg.Po = sg.P

		if mesh.UV.Elems != nil {
			sg.U = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][0] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][0] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][0]
			sg.V = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][1] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][1] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][1]

		} else {
			sg.U = U
			sg.V = V
		}

		sg.DdPdu[0] = e00
		sg.DdPdu[1] = e01
		sg.DdPdu[2] = e02

		sg.DdPdv[0] = e10
		sg.DdPdv[1] = e11
		sg.DdPdv[2] = e12

	} else {
		i0 := int32(idx*3 + 0)
		i1 := int32(idx*3 + 1)
		i2 := int32(idx*3 + 2)

		sg.Shader = mesh.shader

		e00 := mesh.Verts.Elems[i1][0] - mesh.Verts.Elems[i0][0]
		e01 := mesh.Verts.Elems[i1][1] - mesh.Verts.Elems[i0][1]
		e02 := mesh.Verts.Elems[i1][2] - mesh.Verts.Elems[i0][2]
		e10 := mesh.Verts.Elems[i2][0] - mesh.Verts.Elems[i0][0]
		e11 := mesh.Verts.Elems[i2][1] - mesh.Verts.Elems[i0][1]
		e12 := mesh.Verts.Elems[i2][2] - mesh.Verts.Elems[i0][2]

		//	e0 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+1], mesh.Verts.Elems[(idx*3)+0])
		//	e1 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+2], mesh.Verts.Elems[(idx*3)+0])
		//	N := m.Vec3Cross(e0, e1)
		var N m.Vec3

		N[0] = e01*e12 - e02*e11
		N[1] = e02*e10 - e00*e12
		N[2] = e00*e11 - e01*e10
		N = m.Vec3Normalize(N)

		if mesh.Normals.Elems != nil {
			for k := range sg.N {
				sg.N[k] = U*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][k] +
					V*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][k] +
					W*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][k]
			}
			sg.N = m.Vec3Normalize(sg.N)
		} else {
			sg.N = N
		}

		xAbsSum := m.Abs(U*mesh.Verts.Elems[i0][0]) + m.Abs(V*mesh.Verts.Elems[i1][0]) + m.Abs(W*mesh.Verts.Elems[i2][0])
		yAbsSum := m.Abs(U*mesh.Verts.Elems[i0][1]) + m.Abs(V*mesh.Verts.Elems[i1][1]) + m.Abs(W*mesh.Verts.Elems[i2][1])
		zAbsSum := m.Abs(U*mesh.Verts.Elems[i0][2]) + m.Abs(V*mesh.Verts.Elems[i1][2]) + m.Abs(W*mesh.Verts.Elems[i2][2])

		xAbsSum = m.Max(xAbsSum, 0.08)
		yAbsSum = m.Max(yAbsSum, 0.08)
		zAbsSum = m.Max(zAbsSum, 0.08)

		d := m.Gamma(7)*xAbsSum*m.Abs(N[0]) + m.Gamma(7)*yAbsSum*m.Abs(N[1]) + m.Gamma(7)*zAbsSum*m.Abs(N[2])
		offset := m.Vec3Scale(d, N)

		sg.Poffset = offset
		sg.Ng = N
		sg.Bu = U
		sg.Bv = V

		for k := range sg.P {
			sg.P[k] = U*mesh.Verts.Elems[i0][k] +
				V*mesh.Verts.Elems[i1][k] +
				W*mesh.Verts.Elems[i2][k]
		}
		sg.Po = sg.P

		if mesh.UV.Elems != nil {
			sg.U = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][0] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][0] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][0]
			sg.V = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][1] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][1] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][1]

		} else {
			sg.U = U
			sg.V = V
		}

		sg.DdPdu[0] = e00
		sg.DdPdu[1] = e01
		sg.DdPdu[2] = e02

		sg.DdPdv[0] = e10
		sg.DdPdv[1] = e11
		sg.DdPdv[2] = e12

	}
	sg.ElemID = uint32(idx)

	return true

}

// TraceMotionElems implements qbvh.MotionPrimitive.
func (mesh *PolyMesh) TraceMotionElems(time float32, key, key2 int, ray *core.Ray, sg *core.ShaderContext, base, count int) bool {
	return false
}
