// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

// Trace implements core.Primitive.
func (mesh *PolyMesh) Trace(ray *core.Ray, sg *core.ShaderContext) bool {
	// At this point should transform ray.
	var Rp, Rd, Rdinv m.Vec3
	var S [3]float32
	var Kx, Ky, Kz int32
	var transform, invTransform m.Matrix4

	if mesh.Transform.Elems != nil {
		Rp = ray.P
		Rd = ray.D
		Rdinv = ray.Dinv
		S = ray.S
		Kx = ray.Kx
		Ky = ray.Ky
		Kz = ray.Kz

		if len(mesh.Transform.Elems) > 1 {
			k := ray.Time * float32(len(mesh.Transform.Elems)-1)

			time := k - m.Floor(k)

			key := int(m.Floor(k))
			key2 := int(m.Ceil(k))

			if key > len(mesh.Transform.Elems)-1 || key > len(mesh.Transform.Elems)-1 {
				panic(fmt.Sprintf("%v %v %v", ray.Time, key, key2))
			}
			//fmt.Printf("%v %v %v %v %v %v %v", ray.Time, len(mesh.Transform.Elems), time, key, key2, len(mesh.transformSRT), mesh.transformSRT)

			transformSRT := m.TransformDecompLerp(mesh.transformSRT[key], mesh.transformSRT[key2], time)
			transform = m.TransformDecompToMatrix4(transformSRT)
		} else {
			transform = mesh.Transform.Elems[0]
		}

		invTransform, _ = m.Matrix4Inverse(transform)

		ray.P = m.Matrix4MulPoint(invTransform, Rp)
		ray.D = m.Matrix4MulVec(invTransform, Rd)
		ray.Setup()
	}

	if mesh.accel.qbvh != nil {

		// Static
		hit := qbvh.Trace(mesh.accel.qbvh, mesh, ray, sg)

		if mesh.Transform.Elems != nil {
			ray.P = Rp
			ray.D = Rd
			ray.Dinv = Rdinv
			ray.S = S
			ray.Kx = Kx
			ray.Ky = Ky
			ray.Kz = Kz
			sg.Transform = transform
			sg.InvTransform = invTransform
			sg.P = m.Matrix4MulPoint(transform, sg.Po)
			sg.N = m.Matrix4MulVec(m.Matrix4Transpose(invTransform), sg.N)
		}
		return hit
	}

	// Motion
	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))

	hit := qbvh.TraceMotion(mesh.accel.mqbvh, time, key, key2, mesh, ray, sg)

	if mesh.Transform.Elems != nil {
		ray.P = Rp
		ray.D = Rd
		ray.Dinv = Rdinv
		ray.S = S
		ray.Kx = Kx
		ray.Ky = Ky
		ray.Kz = Kz
		sg.Transform = transform
		sg.InvTransform = invTransform
		sg.P = m.Matrix4MulPoint(transform, sg.Po)
		sg.N = m.Matrix4MulVec(m.Matrix4Transpose(transform), sg.N)
	}
	return hit
}

// TraceElems implements qbvh.Primitive.
//go:nosplit
func (mesh *PolyMesh) TraceElems(ray *core.Ray, sg *core.ShaderContext, base, count int) bool {

	// NOTES: by unrolling many of the vector ops we avoid storing vec3's on the stack and allows it to fit
	// within the nosplit stack allowance.
	idx := int32(-1)
	var U, V, W float32

	if mesh.idxp != nil {
		for i := base; i < base+count; i++ {
			//faceidx := mesh.accel.idx[i]

			i0 := int(mesh.idxp[i*3+0])
			i1 := int(mesh.idxp[i*3+1])
			i2 := int(mesh.idxp[i*3+2])

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
			idx = int32(i)
		}
	} else {
		for i := base; i < base+count; i++ {
			faceidx := mesh.accel.idx[i]
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
			idx = faceidx
		}

		// At this point we have an intersection.

	}

	if idx == -1 {
		return false
	}

	var i0, i1, i2 int32

	if mesh.idxp != nil {
		i0 = int32(mesh.idxp[idx*3+0])
		i1 = int32(mesh.idxp[idx*3+1])
		i2 = int32(mesh.idxp[idx*3+2])
	} else {
		i0 = int32(idx*3 + 0)
		i1 = int32(idx*3 + 1)
		i2 = int32(idx*3 + 2)
	}

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

	sg.Ng[0] = e01*e12 - e02*e11
	sg.Ng[1] = e02*e10 - e00*e12
	sg.Ng[2] = e00*e11 - e01*e10
	sg.Ng = m.Vec3Normalize(sg.Ng)

	if mesh.Normals.Elems != nil {
		for k := range sg.N {
			sg.N[k] = U*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][k] +
				V*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][k] +
				W*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][k]
		}

		sg.N = m.Vec3Normalize(sg.N)
	} else {
		sg.N = sg.Ng
	}

	d := m.Gamma(7)*xAbsSum*m.Abs(sg.Ng[0]) + m.Gamma(7)*yAbsSum*m.Abs(sg.Ng[1]) + m.Gamma(7)*zAbsSum*m.Abs(sg.Ng[2])

	sg.Poffset = m.Vec3Scale(d, sg.Ng)
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

	sg.ElemID = uint32(idx)

	return true

}

// TraceMotionElems implements qbvh.MotionPrimitive.
func (mesh *PolyMesh) TraceMotionElems(time float32, key, key2 int, ray *core.Ray, sg *core.ShaderContext, base, count int) bool {
	// NOTES: by unrolling many of the vector ops we avoid storing vec3's on the stack and allows it to fit
	// within the nosplit stack allowance.
	idx := int32(-1)
	var U, V, W float32

	for i := base; i < base+count; i++ {
		//faceidx := mesh.accel.idx[i]
		var i0, i1, i2 int

		faceidx := int32(i)

		if mesh.idxp != nil {
			//faceidx := mesh.accel.idx[i]
			// mesh.idxp is reorderd to match the accel idxs.
			i0 = int(mesh.idxp[i*3+0])
			i1 = int(mesh.idxp[i*3+1])
			i2 = int(mesh.idxp[i*3+2])
		} else {
			// no idxp so faces can't be reordered.
			faceidx = mesh.accel.idx[i]
			i0 = int(faceidx*3 + 0)
			i1 = int(faceidx*3 + 1)
			i2 = int(faceidx*3 + 2)

		}

		V0 := m.Vec3Lerp(mesh.Verts.Elems[i0+mesh.Verts.ElemsPerKey*key],
			mesh.Verts.Elems[i0+mesh.Verts.ElemsPerKey*key2], time)

		V1 := m.Vec3Lerp(mesh.Verts.Elems[i1+mesh.Verts.ElemsPerKey*key],
			mesh.Verts.Elems[i1+mesh.Verts.ElemsPerKey*key2], time)

		V2 := m.Vec3Lerp(mesh.Verts.Elems[i2+mesh.Verts.ElemsPerKey*key],
			mesh.Verts.Elems[i2+mesh.Verts.ElemsPerKey*key2], time)

		AKz := V0[ray.Kz] - ray.P[ray.Kz]
		BKz := V1[ray.Kz] - ray.P[ray.Kz]
		CKz := V2[ray.Kz] - ray.P[ray.Kz]

		var fU, fV, fW float32
		{
			Cx := (V2[ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*CKz
			By := (V1[ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*BKz
			Cy := (V2[ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*CKz
			Bx := (V1[ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*BKz
			Ax := (V0[ray.Kx] - ray.P[ray.Kx]) - ray.S[0]*AKz
			Ay := (V0[ray.Ky] - ray.P[ray.Ky]) - ray.S[1]*AKz

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
		idx = faceidx

		for k := range sg.Po {
			sg.Po[k] = U*V0[k] +
				V*V1[k] +
				W*V2[k]
		}

		xAbsSum := m.Abs(U*V0[0]) + m.Abs(V*V1[0]) + m.Abs(W*V2[0])
		yAbsSum := m.Abs(U*V0[1]) + m.Abs(V*V1[1]) + m.Abs(W*V2[1])
		zAbsSum := m.Abs(U*V0[2]) + m.Abs(V*V1[2]) + m.Abs(W*V2[2])

		//xAbsSum = m.Max(xAbsSum, 0.8)
		//yAbsSum = m.Max(yAbsSum, 0.8)
		//zAbsSum = m.Max(zAbsSum, 0.8)

		e00 := V1[0] - V0[0]
		e01 := V1[1] - V0[1]
		e02 := V1[2] - V0[2]
		e10 := V2[0] - V0[0]
		e11 := V2[1] - V0[1]
		e12 := V2[2] - V0[2]

		//	e0 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+1], mesh.Verts.Elems[(idx*3)+0])
		//	e1 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+2], mesh.Verts.Elems[(idx*3)+0])
		//	N := m.Vec3Cross(e0, e1)

		sg.Ng[0] = e01*e12 - e02*e11
		sg.Ng[1] = e02*e10 - e00*e12
		sg.Ng[2] = e00*e11 - e01*e10
		sg.Ng = m.Vec3Normalize(sg.Ng)

		sg.DdPdu[0] = e00
		sg.DdPdu[1] = e01
		sg.DdPdu[2] = e02

		sg.DdPdv[0] = e10
		sg.DdPdv[1] = e11
		sg.DdPdv[2] = e12

		d := m.Gamma(7)*xAbsSum*m.Abs(sg.Ng[0]) + m.Gamma(7)*yAbsSum*m.Abs(sg.Ng[1]) + m.Gamma(7)*zAbsSum*m.Abs(sg.Ng[2])

		sg.Poffset = m.Vec3Scale(d, sg.Ng)
		sg.Bu = U
		sg.Bv = V

	}

	if idx == -1 {
		return false
	}

	sg.Shader = mesh.shader

	sg.P = sg.Po

	if mesh.UV.Elems != nil {
		sg.U = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][0] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][0] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][0]
		sg.V = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][1] + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][1] + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][1]

	} else {
		sg.U = U
		sg.V = V
	}

	if mesh.Normals.Elems != nil && mesh.Normals.MotionKeys == mesh.Verts.MotionKeys {
		N0 := m.Vec3Lerp(mesh.Normals.Elems[int((idx*3)+0)+mesh.Normals.ElemsPerKey*key],
			mesh.Normals.Elems[int((idx*3)+0)+mesh.Normals.ElemsPerKey*key2], time)

		N1 := m.Vec3Lerp(mesh.Normals.Elems[int((idx*3)+1)+mesh.Normals.ElemsPerKey*key],
			mesh.Normals.Elems[int((idx*3)+1)+mesh.Normals.ElemsPerKey*key2], time)

		N2 := m.Vec3Lerp(mesh.Normals.Elems[int((idx*3)+2)+mesh.Normals.ElemsPerKey*key],
			mesh.Normals.Elems[int((idx*3)+2)+mesh.Normals.ElemsPerKey*key2], time)

		for k := range sg.N {
			sg.N[k] = U*N0[k] +
				V*N1[k] +
				W*N2[k]
		}

		sg.N = m.Vec3Normalize(sg.N)
	} else {
		sg.N = sg.Ng
	}

	sg.ElemID = uint32(idx)

	return true

}
