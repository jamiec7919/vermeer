// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	m "github.com/jamiec7919/vermeer/math"
)

// init should triangulate the polygons and build the idxp list.
func (mesh *PolyMesh) init() error {

	for _, t := range mesh.Transform.Elems {
		mesh.transformSRT = append(mesh.transformSRT, m.Matrix4ToTransform(t))
	}

	mesh.initTransformBounds()

	if mesh.PolyCount != nil {
		basei := uint32(0)
		for k := range mesh.PolyCount {
			i := uint32(0)
			for j := 0; j < int(mesh.PolyCount[k]-2); j++ {
				i++
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[basei]))
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[basei+i]))
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[basei+i+1]))

				V0 := mesh.Verts.Elems[mesh.FaceIdx[basei]]
				V1 := mesh.Verts.Elems[mesh.FaceIdx[basei+i]]
				V2 := mesh.Verts.Elems[mesh.FaceIdx[basei+i+1]]

				if V0 == V1 && V0 == V2 {
					//log.Printf("nil triangle: %v %v %v %v %v\n", mesh.NodeName, mesh.FaceIdx[basei], V0, V1, V2)
				}

				if mesh.UV.Elems != nil {
					if mesh.UVIdx != nil { // if UVIdx doesn't exist assume same as FaceIdx
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[basei]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[basei+i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[basei+i+1]))
					} else {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[basei]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[basei+i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[basei+i+1]))

					}
				}

				if mesh.Normals.Elems != nil {
					if mesh.NormalIdx != nil { // if NormalIdx doesn't exist assume same as FaceIdx
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[basei]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[basei+i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[basei+i+1]))
					} else {
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[basei]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[basei+i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[basei+i+1]))

					}
				}

				if mesh.ShaderIdx != nil {
					mesh.shaderidx = append(mesh.shaderidx, uint8(mesh.ShaderIdx[k]))
				}
			}
			basei += uint32(mesh.PolyCount[k])
		}

	} else {
		// Assume already a triangle mesh
		if mesh.FaceIdx != nil {
			for j := range mesh.FaceIdx {
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[j]))

				if mesh.UV.Elems != nil {
					if mesh.UVIdx != nil {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[j]))
					} else {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[j]))
					}
				}

				if mesh.Normals.Elems != nil {
					if mesh.NormalIdx != nil {
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[j]))
					} else {
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[j]))

					}
				}

			}
		} else {
			// No indexes so assume the vertex array is simply the triangle verts
			for j := 0; j < mesh.Verts.ElemsPerKey; j++ {
				mesh.idxp = append(mesh.idxp, uint32(j))

				if mesh.UV.Elems != nil {
					if mesh.UVIdx != nil {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[j]))
					} else {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(j))

					}
				}

				if mesh.Normals.Elems != nil {
					if mesh.NormalIdx != nil {
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[j]))
					} else {
						mesh.normalidx = append(mesh.normalidx, uint32(j))

					}
				}

			}
		}

		for _, idx := range mesh.ShaderIdx {
			mesh.shaderidx = append(mesh.shaderidx, uint8(idx))
		}

	}

	mesh.FaceIdx = nil
	mesh.PolyCount = nil
	mesh.UVIdx = nil
	mesh.NormalIdx = nil
	mesh.ShaderIdx = nil

	return nil
}
