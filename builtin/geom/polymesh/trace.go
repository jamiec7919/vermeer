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
		sg.N = m.Matrix4MulVec(m.Matrix4Transpose(invTransform), sg.N)
	}
	return hit
}

// TraceElems implements qbvh.Primitive.
///go:nosplit
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
			if m.Xorf(T, detSign) < (m.EpsilonFloat32+mesh.RayBias)*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
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

	shaderIdx := uint8(0)

	if mesh.shaderidx != nil {
		shaderIdx = mesh.shaderidx[idx]
	}

	sg.Shader = mesh.shader[shaderIdx]

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

	// Differential calculation
	// Construct barycentric planes
	//nalphax := m.Vec3Cross(sg.Ng, m.Vec3Sub(mesh.Verts.Elems[i2], mesh.Verts.Elems[i1]))
	nalphax := sg.Ng[1]*(mesh.Verts.Elems[i2][2]-mesh.Verts.Elems[i1][2]) - sg.Ng[2]*(mesh.Verts.Elems[i2][1]-mesh.Verts.Elems[i1][1])
	nalphay := sg.Ng[2]*(mesh.Verts.Elems[i2][0]-mesh.Verts.Elems[i1][0]) - sg.Ng[0]*(mesh.Verts.Elems[i2][2]-mesh.Verts.Elems[i1][2])
	nalphaz := sg.Ng[0]*(mesh.Verts.Elems[i2][1]-mesh.Verts.Elems[i1][1]) - sg.Ng[1]*(mesh.Verts.Elems[i2][0]-mesh.Verts.Elems[i1][0])
	//nalphad := mesh.Verts.Elems[i1][0]*nalphax + mesh.Verts.Elems[i1][1]*nalphay + mesh.Verts.Elems[i1][2]*nalphaz
	nalphad := -mesh.Verts.Elems[i1][0]*nalphax - mesh.Verts.Elems[i1][1]*nalphay - mesh.Verts.Elems[i1][2]*nalphaz

	l := mesh.Verts.Elems[i0][0]*nalphax + mesh.Verts.Elems[i0][1]*nalphay + mesh.Verts.Elems[i0][2]*nalphaz + nalphad

	// Normalize
	nalphax /= l
	nalphay /= l
	nalphaz /= l
	nalphad /= l
	/*
		udot := mesh.Verts.Elems[i1][0]*nalphax + mesh.Verts.Elems[i1][1]*nalphay + mesh.Verts.Elems[i1][2]*nalphaz + nalphad
		vdot := mesh.Verts.Elems[i0][0]*nalphax + mesh.Verts.Elems[i0][1]*nalphay + mesh.Verts.Elems[i0][2]*nalphaz + nalphad
		wdot := mesh.Verts.Elems[i2][0]*nalphax + mesh.Verts.Elems[i2][1]*nalphay + mesh.Verts.Elems[i2][2]*nalphaz + nalphad
		fmt.Printf("%v %v %v\n", udot, vdot, wdot)
	*/
	//l := nalphax*sg.P[0] + nalphay*sg.P[1] + nalphaz*sg.P[2] + nalphad

	//fmt.Printf("a: %v %v %v %v %v %v %v\n", nalphax, nalphay, nalphaz, nalphad, l, sg.Ng, (mesh.Verts.Elems[i2][2] - mesh.Verts.Elems[i1][2]))

	//	nalphad /= l

	nbetax := sg.Ng[1]*(mesh.Verts.Elems[i2][2]-mesh.Verts.Elems[i0][2]) - sg.Ng[2]*(mesh.Verts.Elems[i2][1]-mesh.Verts.Elems[i0][1])
	nbetay := sg.Ng[2]*(mesh.Verts.Elems[i2][0]-mesh.Verts.Elems[i0][0]) - sg.Ng[0]*(mesh.Verts.Elems[i2][2]-mesh.Verts.Elems[i0][2])
	nbetaz := sg.Ng[0]*(mesh.Verts.Elems[i2][1]-mesh.Verts.Elems[i0][1]) - sg.Ng[1]*(mesh.Verts.Elems[i2][0]-mesh.Verts.Elems[i0][0])

	nbetad := -mesh.Verts.Elems[i0][0]*nbetax - mesh.Verts.Elems[i0][1]*nbetay - mesh.Verts.Elems[i0][2]*nbetaz

	l = nbetax*mesh.Verts.Elems[i1][0] + nbetay*mesh.Verts.Elems[i1][1] + nbetaz*mesh.Verts.Elems[i1][2] + nbetad

	// Normalize
	nbetax /= l
	nbetay /= l
	nbetaz /= l
	nbetad /= l

	ngammax := sg.Ng[1]*(mesh.Verts.Elems[i1][2]-mesh.Verts.Elems[i0][2]) - sg.Ng[2]*(mesh.Verts.Elems[i1][1]-mesh.Verts.Elems[i0][1])
	ngammay := sg.Ng[2]*(mesh.Verts.Elems[i1][0]-mesh.Verts.Elems[i0][0]) - sg.Ng[0]*(mesh.Verts.Elems[i1][2]-mesh.Verts.Elems[i0][2])
	ngammaz := sg.Ng[0]*(mesh.Verts.Elems[i1][1]-mesh.Verts.Elems[i0][1]) - sg.Ng[1]*(mesh.Verts.Elems[i1][0]-mesh.Verts.Elems[i0][0])

	ngammad := -mesh.Verts.Elems[i0][0]*ngammax - mesh.Verts.Elems[i0][1]*ngammay - mesh.Verts.Elems[i0][2]*ngammaz

	l = ngammax*mesh.Verts.Elems[i2][0] + ngammay*mesh.Verts.Elems[i2][1] + ngammaz*mesh.Verts.Elems[i2][2] + ngammad

	// Normalize
	ngammax /= l
	ngammay /= l
	ngammaz /= l
	ngammad /= l

	ray.DifferentialTransfer(sg)
	/*
		nbeta := m.Vec3Cross(sg.Ng, m.Vec3Sub(mesh.Verts.Elems[i2], mesh.Verts.Elems[i0]))
		Lbeta := [4]float32{nbeta[0], nbeta[1], nbeta[2], -m.Vec3Dot(mesh.Verts.Elems[i0], nbeta)}

		l = Lbeta[0]*sg.P[0] + Lbeta[1]*sg.P[1] + Lbeta[2]*sg.P[2] + Lbeta[3]

		// Normalize
		Lbeta[0] /= l
		Lbeta[1] /= l
		Lbeta[2] /= l
		Lbeta[3] /= l
	*/

	alphax := nalphax*sg.DdPdx[0] + nalphay*sg.DdPdx[1] + nalphaz*sg.DdPdx[2]
	betax := nbetax*sg.DdPdx[0] + nbetay*sg.DdPdx[1] + nbetaz*sg.DdPdx[2]
	gammax := ngammax*sg.DdPdx[0] + ngammay*sg.DdPdx[1] + ngammaz*sg.DdPdx[2]

	var dndx0, dndx1, dndx2 float32

	if mesh.Normals.Elems != nil {
		//fmt.Printf("%v %v %v\n", alphax, betax, gammax)
		dndx0 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][0] + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][0] + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][0]
		dndx1 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][1] + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][1] + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][1]
		dndx2 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][2] + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][2] + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][2]
	}

	var Ng m.Vec3

	Ng[0] = e01*e12 - e02*e11
	Ng[1] = e02*e10 - e00*e12
	Ng[2] = e00*e11 - e01*e10

	sg.DdNdx = m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(Ng, Ng), m.Vec3{dndx0, dndx1, dndx2}), m.Vec3Scale(m.Vec3Dot(Ng, m.Vec3{dndx0, dndx1, dndx2}), Ng))
	sg.DdNdx = m.Vec3Scale(1/(m.Vec3Dot(Ng, Ng)*m.Sqrt(m.Vec3Dot(Ng, Ng))), sg.DdNdx)

	alphay := nalphax*sg.DdPdy[0] + nalphay*sg.DdPdy[1] + nalphaz*sg.DdPdy[2]
	betay := nbetax*sg.DdPdy[0] + nbetay*sg.DdPdy[1] + nbetaz*sg.DdPdy[2]
	gammay := ngammax*sg.DdPdy[0] + ngammay*sg.DdPdy[1] + ngammaz*sg.DdPdy[2]

	var dndy0, dndy1, dndy2 float32
	if mesh.Normals.Elems != nil {
		dndy0 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][0] + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][0] + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][0]
		dndy1 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][1] + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][1] + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][1]
		dndy2 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]][2] + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]][2] + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]][2]
	}

	sg.DdNdy = m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(Ng, Ng), m.Vec3{dndy0, dndy1, dndy2}), m.Vec3Scale(m.Vec3Dot(Ng, m.Vec3{dndy0, dndy1, dndy2}), Ng))
	sg.DdNdy = m.Vec3Scale(1/(m.Vec3Dot(Ng, Ng)*m.Sqrt(m.Vec3Dot(Ng, Ng))), sg.DdNdy)

	//talpha :=x nalphax*sg.P[0] + nalphay*sg.P[1] + nalphaz*sg.P[2] + nalphad
	//tbeta := nbetax*sg.P[0] + nbetay*sg.P[1] + nbetaz*sg.P[2] + nbetad
	//tgamma := ngammax*sg.P[0] + ngammay*sg.P[1] + ngammaz*sg.P[2] + ngammad

	//fmt.Printf("%v %v %v -> %v %v %v\n", U, V, W, talpha, tbeta, tgamma)
	if mesh.UV.Elems != nil {
		for k := range sg.Dduvdx {
			sg.Dduvdx[k] = alphax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][k] + betax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][k] + gammax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][k]
			sg.Dduvdy[k] = alphay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][k] + betay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][k] + gammay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][k]

		} //	sg.U = talpha*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][0] + tbeta*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][0] + tgamma*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][0]
		//	sg.V = talpha*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]][1] + tbeta*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]][1] + tgamma*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]][1]
	} else {
		sg.Dduvdx[0] = alphax*0 + betax*1 + gammax*0
		sg.Dduvdx[1] = alphax*0 + betax*0 + gammax*1

		sg.Dduvdy[0] = alphay*0 + betay*1 + gammay*0
		sg.Dduvdy[1] = alphay*0 + betay*0 + gammay*1
	}
	//fmt.Printf("%v %v \n", sg.Dduvdx, sg.Dduvdy)
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

	shaderIdx := uint8(0)

	if mesh.shaderidx != nil {
		shaderIdx = mesh.shaderidx[idx]
	}

	sg.Shader = mesh.shader[shaderIdx]

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
