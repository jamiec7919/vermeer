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
	transform := m.Matrix4Identity()
	invTransform := m.Matrix4Identity()

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

			transformSRT := m.TransformLerp(mesh.transformSRT[key], mesh.transformSRT[key2], time)
			transform = m.TransformToMatrix4(transformSRT)
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

			if hit {
				sg.Transform = transform
				sg.InvTransform = invTransform
				sg.P = m.Matrix4MulPoint(transform, sg.Po)
			}
			//sg.N = m.Matrix4MulVec(m.Matrix4Transpose(invTransform), sg.N)
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

		if hit {
			sg.Transform = transform
			sg.InvTransform = invTransform
			sg.P = m.Matrix4MulPoint(transform, sg.Po)
			//sg.N = m.Matrix4MulVec(m.Matrix4Transpose(invTransform), sg.N)
		}
	}
	return hit
}

func sqr(x float32) float32 { return x * x }

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

			Kz := int(ray.Kz)

			AKz := mesh.Verts.Elems[i0].Elt(Kz) - ray.P.Elt(Kz)
			BKz := mesh.Verts.Elems[i1].Elt(Kz) - ray.P.Elt(Kz)
			CKz := mesh.Verts.Elems[i2].Elt(Kz) - ray.P.Elt(Kz)

			var fU, fV, fW float32
			{
				Kx := int(ray.Kx)
				Ky := int(ray.Ky)
				Cx := (mesh.Verts.Elems[i2].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*CKz
				By := (mesh.Verts.Elems[i1].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*BKz
				Cy := (mesh.Verts.Elems[i2].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*CKz
				Bx := (mesh.Verts.Elems[i1].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*BKz
				Ax := (mesh.Verts.Elems[i0].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*AKz
				Ay := (mesh.Verts.Elems[i0].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*AKz

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
			if m.Xorf(T, detSign) <= (m.EpsilonFloat32+mesh.RayBias+ray.Tmin)*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
				continue
			}

			rcpDet := 1.0 / det

			//if ray.Tmin >= T*rcpDet {
			//	continue
			//}

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

			Kz := int(ray.Kz)

			AKz := mesh.Verts.Elems[i0].Elt(Kz) - ray.P.Elt(Kz)
			BKz := mesh.Verts.Elems[i1].Elt(Kz) - ray.P.Elt(Kz)
			CKz := mesh.Verts.Elems[i2].Elt(Kz) - ray.P.Elt(Kz)

			var fU, fV, fW float32
			{
				Kx := int(ray.Kx)
				Ky := int(ray.Ky)
				Cx := (mesh.Verts.Elems[i2].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*CKz
				By := (mesh.Verts.Elems[i1].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*BKz
				Cy := (mesh.Verts.Elems[i2].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*CKz
				Bx := (mesh.Verts.Elems[i1].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*BKz
				Ax := (mesh.Verts.Elems[i0].Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*AKz
				Ay := (mesh.Verts.Elems[i0].Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*AKz

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

			if m.Xorf(T, detSign) <= (m.EpsilonFloat32+mesh.RayBias+ray.Tmin)*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
				continue
			}

			rcpDet := 1.0 / det

			//	if ray.Tmin >= T*rcpDet {
			//		continue
			//	}

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

	xAbsSum := m.Abs(U*mesh.Verts.Elems[i0].X) + m.Abs(V*mesh.Verts.Elems[i1].X) + m.Abs(W*mesh.Verts.Elems[i2].X)
	yAbsSum := m.Abs(U*mesh.Verts.Elems[i0].Y) + m.Abs(V*mesh.Verts.Elems[i1].Y) + m.Abs(W*mesh.Verts.Elems[i2].Y)
	zAbsSum := m.Abs(U*mesh.Verts.Elems[i0].Z) + m.Abs(V*mesh.Verts.Elems[i1].Z) + m.Abs(W*mesh.Verts.Elems[i2].Z)

	//xAbsSum = m.Max(xAbsSum, 0.8)
	//yAbsSum = m.Max(yAbsSum, 0.8)
	//zAbsSum = m.Max(zAbsSum, 0.8)

	shaderIdx := uint8(0)

	if mesh.shaderidx != nil {
		shaderIdx = mesh.shaderidx[idx]
	}

	if mesh.shader == nil || int(shaderIdx) >= len(mesh.shader) {
		panic(fmt.Sprintf("%v: Shader %v %v", mesh.Name(), mesh.Shader, shaderIdx))
	}

	sg.Shader = mesh.shader[shaderIdx]

	e00 := mesh.Verts.Elems[i1].X - mesh.Verts.Elems[i0].X
	e01 := mesh.Verts.Elems[i1].Y - mesh.Verts.Elems[i0].Y
	e02 := mesh.Verts.Elems[i1].Z - mesh.Verts.Elems[i0].Z
	e10 := mesh.Verts.Elems[i2].X - mesh.Verts.Elems[i0].X
	e11 := mesh.Verts.Elems[i2].Y - mesh.Verts.Elems[i0].Y
	e12 := mesh.Verts.Elems[i2].Z - mesh.Verts.Elems[i0].Z

	//	e0 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+1], mesh.Verts.Elems[(idx*3)+0])
	//	e1 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+2], mesh.Verts.Elems[(idx*3)+0])
	//	N := m.Vec3Cross(e0, e1)

	sg.Ng.X = e01*e12 - e02*e11
	sg.Ng.Y = e02*e10 - e00*e12
	sg.Ng.Z = e00*e11 - e01*e10
	sg.Ng = m.Vec3Normalize(sg.Ng)

	var N m.Vec3

	if mesh.Normals.Elems != nil {
		for k := 0; k < 3; k++ {
			sg.N.Set(k, U*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].Elt(k)+
				V*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].Elt(k)+
				W*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].Elt(k))
		}

		N = sg.N
		sg.N = m.Vec3Normalize(sg.N)
	} else {
		N = sg.Ng
		sg.N = sg.Ng
	}

	d := m.Gamma(7)*xAbsSum*m.Abs(sg.Ng.X) + m.Gamma(7)*yAbsSum*m.Abs(sg.Ng.Y) + m.Gamma(7)*zAbsSum*m.Abs(sg.Ng.Z)

	sg.Poffset = m.Vec3Scale(d, sg.Ng)
	sg.Bu = U
	sg.Bv = V

	for k := 0; k < 3; k++ {
		sg.P.Set(k, U*mesh.Verts.Elems[i0].Elt(k)+
			V*mesh.Verts.Elems[i1].Elt(k)+
			W*mesh.Verts.Elems[i2].Elt(k))
	}
	sg.Po = sg.P

	if mesh.UV.Elems != nil && mesh.uvtriidx != nil {
		sg.U = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].X + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].X + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].X
		sg.V = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].Y + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].Y + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].Y
		//sg.U = U*mesh.UV.Elems[mesh.uvtriidx[i0]].X + V*mesh.UV.Elems[mesh.uvtriidx[i1]].X + W*mesh.UV.Elems[mesh.uvtriidx[i2]].X
		//sg.V = U*mesh.UV.Elems[mesh.uvtriidx[i0]].Y + V*mesh.UV.Elems[mesh.uvtriidx[i1]].Y + W*mesh.UV.Elems[mesh.uvtriidx[i2]].Y

	} else {
		sg.U = U
		sg.V = V
	}

	ray.DifferentialTransfer(sg)

	// Differential calculation
	// Construct barycentric planes
	//nalphax := m.Vec3Cross(sg.Ng, m.Vec3Sub(mesh.Verts.Elems[i2], mesh.Verts.Elems[i1]))
	nalphax := sg.Ng.Y*(mesh.Verts.Elems[i2].Z-mesh.Verts.Elems[i1].Z) - sg.Ng.Z*(mesh.Verts.Elems[i2].Y-mesh.Verts.Elems[i1].Y)
	nalphay := sg.Ng.Z*(mesh.Verts.Elems[i2].X-mesh.Verts.Elems[i1].X) - sg.Ng.X*(mesh.Verts.Elems[i2].Z-mesh.Verts.Elems[i1].Z)
	nalphaz := sg.Ng.X*(mesh.Verts.Elems[i2].Y-mesh.Verts.Elems[i1].Y) - sg.Ng.Y*(mesh.Verts.Elems[i2].X-mesh.Verts.Elems[i1].X)
	//nalphad := mesh.Verts.Elems[i1].X*nalphax + mesh.Verts.Elems[i1].Y*nalphay + mesh.Verts.Elems[i1].Z*nalphaz
	q := m.Sqrt(sqr(nalphax) + sqr(nalphay) + sqr(nalphaz))
	nalphax /= q
	nalphay /= q
	nalphaz /= q

	nalphad := -mesh.Verts.Elems[i1].X*nalphax - mesh.Verts.Elems[i1].Y*nalphay - mesh.Verts.Elems[i1].Z*nalphaz

	l := mesh.Verts.Elems[i0].X*nalphax + mesh.Verts.Elems[i0].Y*nalphay + mesh.Verts.Elems[i0].Z*nalphaz + nalphad

	// Normalize
	nalphax /= l
	nalphay /= l
	nalphaz /= l
	nalphad /= l

	//l := nalphax*sg.P.X + nalphay*sg.P.Y + nalphaz*sg.P.Z + nalphad

	//fmt.Printf("a: %v %v %v %v %v %v %v\n", nalphax, nalphay, nalphaz, nalphad, l, sg.Ng, (mesh.Verts.Elems[i2].Z - mesh.Verts.Elems[i1].Z))

	//	nalphad /= l

	nbetax := sg.Ng.Y*(mesh.Verts.Elems[i2].Z-mesh.Verts.Elems[i0].Z) - sg.Ng.Z*(mesh.Verts.Elems[i2].Y-mesh.Verts.Elems[i0].Y)
	nbetay := sg.Ng.Z*(mesh.Verts.Elems[i2].X-mesh.Verts.Elems[i0].X) - sg.Ng.X*(mesh.Verts.Elems[i2].Z-mesh.Verts.Elems[i0].Z)
	nbetaz := sg.Ng.X*(mesh.Verts.Elems[i2].Y-mesh.Verts.Elems[i0].Y) - sg.Ng.Y*(mesh.Verts.Elems[i2].X-mesh.Verts.Elems[i0].X)
	q = m.Sqrt(sqr(nbetax) + sqr(nbetay) + sqr(nbetaz))
	nbetax /= q
	nbetay /= q
	nbetaz /= q

	nbetad := -mesh.Verts.Elems[i0].X*nbetax - mesh.Verts.Elems[i0].Y*nbetay - mesh.Verts.Elems[i0].Z*nbetaz

	l = nbetax*mesh.Verts.Elems[i1].X + nbetay*mesh.Verts.Elems[i1].Y + nbetaz*mesh.Verts.Elems[i1].Z + nbetad

	// Normalize
	nbetax /= l
	nbetay /= l
	nbetaz /= l
	nbetad /= l

	ngammax := sg.Ng.Y*(mesh.Verts.Elems[i1].Z-mesh.Verts.Elems[i0].Z) - sg.Ng.Z*(mesh.Verts.Elems[i1].Y-mesh.Verts.Elems[i0].Y)
	ngammay := sg.Ng.Z*(mesh.Verts.Elems[i1].X-mesh.Verts.Elems[i0].X) - sg.Ng.X*(mesh.Verts.Elems[i1].Z-mesh.Verts.Elems[i0].Z)
	ngammaz := sg.Ng.X*(mesh.Verts.Elems[i1].Y-mesh.Verts.Elems[i0].Y) - sg.Ng.Y*(mesh.Verts.Elems[i1].X-mesh.Verts.Elems[i0].X)

	q = m.Sqrt(sqr(ngammax) + sqr(ngammay) + sqr(ngammaz))
	ngammax /= q
	ngammay /= q
	ngammaz /= q

	ngammad := -mesh.Verts.Elems[i0].X*ngammax - mesh.Verts.Elems[i0].Y*ngammay - mesh.Verts.Elems[i0].Z*ngammaz

	l = ngammax*mesh.Verts.Elems[i2].X + ngammay*mesh.Verts.Elems[i2].Y + ngammaz*mesh.Verts.Elems[i2].Z + ngammad

	// Normalize
	ngammax /= l
	ngammay /= l
	ngammaz /= l
	ngammad /= l
	/*
		udot := sg.P.X*nalphax + sg.P.Y*nalphay + sg.P.Z*nalphaz + nalphad
		vdot := sg.P.X*nbetax + sg.P.Y*nbetay + sg.P.Z*nbetaz + nbetad
		wdot := sg.P.X*ngammax + sg.P.Y*ngammay + sg.P.Z*ngammaz + ngammad
		fmt.Printf("%v %v %v > %v %v %v\n", udot, vdot, wdot, U, V, W)
	*/
	/*
		nbeta := m.Vec3Cross(sg.Ng, m.Vec3Sub(mesh.Verts.Elems[i2], mesh.Verts.Elems[i0]))
		Lbeta := [4]float32{nbeta.X, nbeta.Y, nbeta.Z, -m.Vec3Dot(mesh.Verts.Elems[i0], nbeta)}

		l = Lbeta.X*sg.P.X + Lbeta.Y*sg.P.Y + Lbeta.Z*sg.P.Z + Lbeta[3]

		// Normalize
		Lbeta.X /= l
		Lbeta.Y /= l
		Lbeta.Z /= l
		Lbeta[3] /= l
	*/

	alphax := nalphax*sg.DdPdx.X + nalphay*sg.DdPdx.Y + nalphaz*sg.DdPdx.Z
	betax := nbetax*sg.DdPdx.X + nbetay*sg.DdPdx.Y + nbetaz*sg.DdPdx.Z
	gammax := ngammax*sg.DdPdx.X + ngammay*sg.DdPdx.Y + ngammaz*sg.DdPdx.Z

	var dndx0, dndx1, dndx2 float32

	if mesh.Normals.Elems != nil {
		//fmt.Printf("%v %v %v\n", alphax, betax, gammax)
		dndx0 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].X + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].X + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].X
		dndx1 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].Y + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].Y + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].Y
		dndx2 = alphax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].Z + betax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].Z + gammax*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].Z
	}

	var Ng m.Vec3

	Ng.X = e01*e12 - e02*e11
	Ng.Y = e02*e10 - e00*e12
	Ng.Z = e00*e11 - e01*e10
	Ng = m.Vec3Normalize(Ng)

	sg.DdNdx = m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(N, N), m.Vec3{dndx0, dndx1, dndx2}), m.Vec3Scale(m.Vec3Dot(N, m.Vec3{dndx0, dndx1, dndx2}), N))
	sg.DdNdx = m.Vec3Scale(1/(m.Vec3Dot(N, N)*m.Sqrt(m.Vec3Dot(N, N))), sg.DdNdx)

	alphay := nalphax*sg.DdPdy.X + nalphay*sg.DdPdy.Y + nalphaz*sg.DdPdy.Z
	betay := nbetax*sg.DdPdy.X + nbetay*sg.DdPdy.Y + nbetaz*sg.DdPdy.Z
	gammay := ngammax*sg.DdPdy.X + ngammay*sg.DdPdy.Y + ngammaz*sg.DdPdy.Z

	var dndy0, dndy1, dndy2 float32
	if mesh.Normals.Elems != nil {
		dndy0 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].X + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].X + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].X
		dndy1 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].Y + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].Y + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].Y
		dndy2 = alphay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+0]].Z + betay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+1]].Z + gammay*mesh.Normals.Elems[mesh.normalidx[(idx*3)+2]].Z
	}

	sg.DdNdy = m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(N, N), m.Vec3{dndy0, dndy1, dndy2}), m.Vec3Scale(m.Vec3Dot(N, m.Vec3{dndy0, dndy1, dndy2}), N))
	sg.DdNdy = m.Vec3Scale(1/(m.Vec3Dot(N, N)*m.Sqrt(m.Vec3Dot(N, N))), sg.DdNdy)

	//talpha :=x nalphax*sg.P.X + nalphay*sg.P.Y + nalphaz*sg.P.Z + nalphad
	//tbeta := nbetax*sg.P.X + nbetay*sg.P.Y + nbetaz*sg.P.Z + nbetad
	//tgamma := ngammax*sg.P.X + ngammay*sg.P.Y + ngammaz*sg.P.Z + ngammad

	//fmt.Printf("%v %v %v -> %v %v %v\n", U, V, W, talpha, tbeta, tgamma)
	if mesh.UV.Elems != nil {
		for k := 0; k < 2; k++ {
			sg.Dduvdx.Set(k, alphax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].Elt(k)+betax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].Elt(k)+gammax*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].Elt(k))
			sg.Dduvdy.Set(k, alphay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].Elt(k)+betay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].Elt(k)+gammay*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].Elt(k))

		} //	sg.U = talpha*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].X + tbeta*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].X + tgamma*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].X
		//	sg.V = talpha*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].Y + tbeta*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].Y + tgamma*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].Y
	} else {
		sg.Dduvdx.X = alphax*0 + betax*1 + gammax*0
		sg.Dduvdx.Y = alphax*0 + betax*0 + gammax*1

		sg.Dduvdy.X = alphay*0 + betay*1 + gammay*0
		sg.Dduvdy.Y = alphay*0 + betay*0 + gammay*1
	}

	axisu := m.Vec3Sub(m.Vec3{1, 0, 0}, m.Vec3Scale(m.Vec3Dot(m.Vec3{1, 0, 0}, sg.Ng), sg.Ng))

	if m.Vec3Length2(axisu) < 0.1 || m.Abs(m.Vec3Dot(axisu, sg.Ng)) > 0.3 {
		axisu = m.Vec3Sub(m.Vec3{0, 0, 1}, m.Vec3Scale(m.Vec3Dot(m.Vec3{0, 0, 1}, sg.Ng), sg.Ng))
	}

	sg.DdPdu = m.Vec3Normalize(axisu)
	sg.DdPdv = m.Vec3Cross(sg.Ng, sg.DdPdu)

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

		Kz := int(ray.Kz)

		AKz := V0.Elt(Kz) - ray.P.Elt(Kz)
		BKz := V1.Elt(Kz) - ray.P.Elt(Kz)
		CKz := V2.Elt(Kz) - ray.P.Elt(Kz)

		var fU, fV, fW float32
		{
			Kx := int(ray.Kx)
			Ky := int(ray.Ky)
			Cx := (V2.Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*CKz
			By := (V1.Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*BKz
			Cy := (V2.Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*CKz
			Bx := (V1.Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*BKz
			Ax := (V0.Elt(Kx) - ray.P.Elt(Kx)) - ray.S[0]*AKz
			Ay := (V0.Elt(Ky) - ray.P.Elt(Ky)) - ray.S[1]*AKz

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

		if m.Xorf(T, detSign) < (m.EpsilonFloat32+mesh.RayBias+ray.Tmin)*m.Xorf(det, detSign) || m.Xorf(T, detSign) > ray.Tclosest*m.Xorf(det, detSign) {
			continue
		}

		rcpDet := 1.0 / det

		U = fU * rcpDet
		V = fV * rcpDet
		W = fW * rcpDet

		ray.Tclosest = T * rcpDet
		idx = faceidx

		for k := 0; k < 3; k++ {
			sg.Po.Set(k, U*V0.Elt(k)+
				V*V1.Elt(k)+
				W*V2.Elt(k))
		}

		xAbsSum := m.Abs(U*V0.X) + m.Abs(V*V1.X) + m.Abs(W*V2.X)
		yAbsSum := m.Abs(U*V0.Y) + m.Abs(V*V1.Y) + m.Abs(W*V2.Y)
		zAbsSum := m.Abs(U*V0.Z) + m.Abs(V*V1.Z) + m.Abs(W*V2.Z)

		//xAbsSum = m.Max(xAbsSum, 0.8)
		//yAbsSum = m.Max(yAbsSum, 0.8)
		//zAbsSum = m.Max(zAbsSum, 0.8)

		e00 := V1.X - V0.X
		e01 := V1.Y - V0.Y
		e02 := V1.Z - V0.Z
		e10 := V2.X - V0.X
		e11 := V2.Y - V0.Y
		e12 := V2.Z - V0.Z

		//	e0 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+1], mesh.Verts.Elems[(idx*3)+0])
		//	e1 := m.Vec3Sub(mesh.Verts.Elems[(idx*3)+2], mesh.Verts.Elems[(idx*3)+0])
		//	N := m.Vec3Cross(e0, e1)

		sg.Ng.X = e01*e12 - e02*e11
		sg.Ng.Y = e02*e10 - e00*e12
		sg.Ng.Z = e00*e11 - e01*e10
		sg.Ng = m.Vec3Normalize(sg.Ng)

		sg.DdPdu.X = e00
		sg.DdPdu.Y = e01
		sg.DdPdu.Z = e02

		sg.DdPdv.X = e10
		sg.DdPdv.Y = e11
		sg.DdPdv.Z = e12

		d := m.Gamma(7)*xAbsSum*m.Abs(sg.Ng.X) + m.Gamma(7)*yAbsSum*m.Abs(sg.Ng.Y) + m.Gamma(7)*zAbsSum*m.Abs(sg.Ng.Z)

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
		sg.U = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].X + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].X + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].X
		sg.V = U*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+0]].Y + V*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+1]].Y + W*mesh.UV.Elems[mesh.uvtriidx[(idx*3)+2]].Y

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

		for k := 0; k < 3; k++ {
			sg.N.Set(k, U*N0.Elt(k)+
				V*N1.Elt(k)+
				W*N2.Elt(k))
		}

		sg.N = m.Vec3Normalize(sg.N)
	} else {
		sg.N = sg.Ng
	}

	sg.ElemID = uint32(idx)

	return true

}
