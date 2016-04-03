// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mesh

import (
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
)

type facevert struct {
	Ni, UVi int32 // Indexes
}

type PolyMesh struct {
	NodeName     string `node:"Name"`
	ModelToWorld m.Matrix4

	Verts     []m.Vec3
	Normals   []m.Vec3
	FaceIdx   []int32
	NormalIdx []int32
	PolyCount []int32
	UV_0      []m.Vec2
	UVIdx_0   []int32

	Material    string
	CalcNormals bool

	faces     []FaceGeom
	normals   []m.Vec3
	uv_0      []m.Vec2
	faceverts map[facevert]int32 // Mapping from facevert to UV+Normal indexes

	mesh *Mesh
}

func (mesh *PolyMesh) WorldBounds() (out m.BoundingBox) {

	return mesh.mesh.WorldBounds()
}

func (mesh *PolyMesh) Name() string { return mesh.NodeName }
func (mesh *PolyMesh) PreRender(rc *core.RenderContext) error {
	if err := mesh.createMesh(rc); err != nil {
		return err
	}

	mesh.mesh.initFaces()

	if mesh.CalcNormals {
		if err := mesh.mesh.calcVertexNormals(); err != nil {
			return err
		}
	}

	return mesh.mesh.initAccel()
}
func (mesh *PolyMesh) PostRender(rc *core.RenderContext) error { return nil }

func (mesh *PolyMesh) TraceRay(ray *core.RayData) {
	mesh.TraceRay(ray)
}

func (mesh *PolyMesh) triangulateFace(base, count int32, mtlid material.Id) {

	for i := int32(0); i < count-2; i++ {
		var face FaceGeom

		face.V[0] = mesh.Verts[mesh.FaceIdx[base]]
		face.V[1] = mesh.Verts[mesh.FaceIdx[base+i+1]]
		face.V[2] = mesh.Verts[mesh.FaceIdx[base+i+2]]

		Ni := [3]int32{-1, -1, -1}
		if mesh.Normals != nil {
			if mesh.NormalIdx != nil {
				Ni[0] = mesh.NormalIdx[base]
				Ni[1] = mesh.NormalIdx[base+i+1]
				Ni[2] = mesh.NormalIdx[base+i+2]

			} else {
				Ni[0] = mesh.FaceIdx[base]
				Ni[1] = mesh.FaceIdx[base+i+1]
				Ni[2] = mesh.FaceIdx[base+i+2]

			}
		}

		UVi := [3]int32{-1, -1, -1}
		if mesh.UV_0 != nil {
			if mesh.UVIdx_0 != nil {
				UVi[0] = mesh.UVIdx_0[base]
				UVi[1] = mesh.UVIdx_0[base+i+1]
				UVi[2] = mesh.UVIdx_0[base+i+2]

			} else {
				UVi[0] = mesh.FaceIdx[base]
				UVi[1] = mesh.FaceIdx[base+i+1]
				UVi[2] = mesh.FaceIdx[base+i+2]

			}
		}

		for k := 0; k < 3; k++ {
			fv := facevert{Ni[k], UVi[k]}

			if fvi, present := mesh.faceverts[fv]; present {
				face.Vi[k] = fvi
			} else {
				// allocate normal and uv
				if len(mesh.normals) > len(mesh.uv_0) {
					fvi = int32(len(mesh.normals))

				} else {
					fvi = int32(len(mesh.uv_0))
				}

				mesh.faceverts[fv] = fvi

				face.Vi[k] = fvi

				if fv.Ni != -1 {
					mesh.normals = append(mesh.normals, mesh.Normals[fv.Ni])
				}
				if fv.UVi != -1 {
					mesh.uv_0 = append(mesh.uv_0, mesh.UV_0[fv.UVi])
				}

			}
		}

		face.MtlId = int32(mtlid)

		mesh.faces = append(mesh.faces, face)
	}
}

func (mesh *PolyMesh) createMesh(rc *core.RenderContext) error {
	idx := int32(0)

	mesh.faceverts = make(map[facevert]int32)

	mtlid := rc.GetMaterialId(mesh.Material)

	for i := range mesh.PolyCount {
		mesh.triangulateFace(idx, mesh.PolyCount[i], mtlid)
		idx += mesh.PolyCount[i]
	}

	mesh.faceverts = nil

	mesh.mesh = &Mesh{}
	mesh.mesh.Faces = mesh.faces

	if mesh.normals != nil {
		mesh.mesh.Vn = mesh.normals
	}

	if mesh.uv_0 != nil {
		mesh.mesh.Vuv = [][]m.Vec2{mesh.uv_0}
	}

	mesh.mesh.Transform(mesh.ModelToWorld)

	return nil
}

func (mesh *PolyMesh) VisRay(ray *core.RayData) {
	mesh.mesh.VisRay(ray)
}

func init() {
	core.RegisterType("PolyMesh", func(rc *core.RenderContext, params core.Params) (interface{}, error) {
		mesh := PolyMesh{ModelToWorld: m.Matrix4Identity}

		if err := params.Unmarshal(&mesh); err != nil {
			return nil, err
		}

		return &mesh, nil
	})
}
