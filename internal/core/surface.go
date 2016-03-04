// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
)

type SurfacePoint struct {
	P       m.Vec3
	N, T, B m.Vec3
	Ns      m.Vec3
	pOffset m.Vec3    // Offset to make sure any intersection point is outside face
	MtlId   int32     // 64 bytes (first line)
	UV      [6]m.Vec2 // 12 floats (48 bytes)
	Extra   map[string]interface{}
}

func (surf *SurfacePoint) TangentToWorld(v m.Vec3) m.Vec3 {
	return m.Vec3BasisExpand(surf.T, surf.B, surf.N, v)
}

func (surf *SurfacePoint) WorldToTangent(v m.Vec3) m.Vec3 {
	return m.Vec3BasisProject(surf.T, surf.B, surf.N, v)
}
