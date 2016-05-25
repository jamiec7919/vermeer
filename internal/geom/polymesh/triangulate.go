// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

// init should triangulate the polygons and build the idxp list.
func (mesh *PolyMesh) init() error {

	if mesh.PolyCount != nil {
		i := uint32(0)
		for k := range mesh.PolyCount {
			base_i := i
			for j := 0; j < int(mesh.PolyCount[k]-2); j++ {
				i++
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[base_i]))
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[i]))
				mesh.idxp = append(mesh.idxp, uint32(mesh.FaceIdx[i+1]))

				if mesh.UV.Elems != nil {
					if mesh.UVIdx != nil { // if UVIdx doesn't exist assume same as FaceIdx
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[base_i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.UVIdx[i+1]))
					} else {
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[base_i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[i]))
						mesh.uvtriidx = append(mesh.uvtriidx, uint32(mesh.FaceIdx[i+1]))

					}
				}

				if mesh.Normals.Elems != nil {
					if mesh.NormalIdx != nil { // if NormalIdx doesn't exist assume same as FaceIdx
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[base_i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.NormalIdx[i+1]))
					} else {
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[base_i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[i]))
						mesh.normalidx = append(mesh.normalidx, uint32(mesh.FaceIdx[i+1]))

					}
				}

			}
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
	}

	mesh.FaceIdx = nil
	mesh.PolyCount = nil
	mesh.UVIdx = nil
	mesh.NormalIdx = nil

	return nil
}
