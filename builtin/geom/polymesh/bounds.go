// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	m "github.com/jamiec7919/vermeer/math"
)

func (mesh *PolyMesh) initTransformBounds() {
	for _, t := range mesh.Transform.Elems {
		box := m.BoundingBox{}
		box.Reset()

		for i := range mesh.Verts.Elems {
			box.GrowVec3(m.Matrix4MulPoint(t, mesh.Verts.Elems[i]))
		}

		mesh.transformBounds = append(mesh.transformBounds, box)
	}
}

// Bounds implements core.Geom.
func (mesh *PolyMesh) Bounds(time float32) m.BoundingBox {
	if false && mesh.transformBounds != nil {
		k := time * float32(len(mesh.transformBounds)-1)

		t := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))

		return m.BoundingBoxLerp(mesh.transformBounds[key], mesh.transformBounds[key2], t)

	}

	if mesh.accel.qbvh != nil {
		// Static
		return mesh.bounds
	}

	// Motion
	k := time * float32(len(mesh.motionBounds)-1)

	t := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))

	return m.BoundingBoxLerp(mesh.motionBounds[key], mesh.motionBounds[key2], t)

}
