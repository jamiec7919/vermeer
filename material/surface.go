// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	m "github.com/jamiec7919/vermeer/math"
)

type SurfacePoint struct {
	P       m.Vec3
	N, T, B m.Vec3
	Ns      m.Vec3
	POffset m.Vec3    // Offset to make sure any intersection point is outside face
	MtlId   int32     // 64 bytes (first line)
	UV      [4]m.Vec2 // 12 floats (48 bytes)
	Pu, Pv  [4]m.Vec3
	Extra   map[string]interface{}
}

func (surf *SurfacePoint) TangentToWorld(v m.Vec3) m.Vec3 {
	return m.Vec3BasisExpand(surf.T, surf.B, surf.Ns, v)
}

func (surf *SurfacePoint) WorldToTangent(v m.Vec3) m.Vec3 {
	return m.Vec3BasisProject(surf.T, surf.B, surf.Ns, v)
}

// The texture derivatives Pu & Pv should already be computed.
func (surf *SurfacePoint) SetupTangentSpace(Ns m.Vec3) {
	surf.Ns = m.Vec3Normalize(Ns)
	surf.B = m.Vec3Normalize(m.Vec3Cross(surf.Ns, surf.Pu[0]))
	surf.T = m.Vec3Normalize(m.Vec3Cross(surf.Ns, surf.B))

}

// Offset the surface point out from surface by about 1ulp
// Pass -ve value to push point 'into' surface (for transmission)
func (r *SurfacePoint) OffsetP(dir int) {

	if dir < 0 {
		r.POffset = m.Vec3Neg(r.POffset)

	}
	po := m.Vec3Add(r.P, r.POffset)

	// round po away from p
	for i := range po {
		//log.Printf("%v %v %v", i, offset[i], po[i])
		if r.POffset[i] > 0 {
			po[i] = m.NextFloatUp(po[i])
		} else if r.POffset[i] < 0 {
			po[i] = m.NextFloatDown(po[i])
		}
		//log.Printf("%v %v", i, po[i])
	}

	r.P = po
}
