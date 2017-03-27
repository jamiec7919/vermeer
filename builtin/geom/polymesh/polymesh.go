// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
	"github.com/jamiec7919/vermeer/qbvh"
)

// PolyMesh is the main polygon mesh type in Vermeer.
type PolyMesh struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	RayBias  float32      `node:",opt"`

	Verts     param.PointArray
	PolyCount []int32 `node:",opt"`
	FaceIdx   []int32 `node:",opt"`

	Shader    []string
	ShaderIdx []int32 `node:",opt"`

	CalcNormals bool `node:",opt"`
	IsVisible   bool `node:",opt"`

	Transform    param.MatrixArray `node:",opt"`
	transformSRT []m.TransformDecomp
	//	invTransformSRT []m.TransformDecomp

	UV    param.Vec2Array `node:",opt"`
	UVIdx []int32         `node:",opt"`

	Normals   param.Vec3Array `node:",opt"`
	NormalIdx []int32         `node:",opt"`

	facecount     int      // Number of faces
	idxp          []uint32 // Triangle Face indexes (position)
	vertidxstride int      // 3 or 4 if including material ids
	uvtriidx      []uint32 // triangulated UV indexes
	normalidx     []uint32
	shaderidx     []uint8

	accel struct {
		mqbvh qbvh.MotionQBVH
		qbvh  []qbvh.Node
		idx   []int32 // Face indexes
	}

	shader []core.Shader

	bounds          m.BoundingBox
	motionBounds    []m.BoundingBox
	transformBounds []m.BoundingBox
}

// Assert that PolyMesh implements important interfaces.
var _ core.Node = (*PolyMesh)(nil)
var _ core.Geom = (*PolyMesh)(nil)

// Name is a core.Node method.
func (mesh *PolyMesh) Name() string { return mesh.NodeName }

// Def is a core.Node method.
func (mesh *PolyMesh) Def() core.NodeDef { return mesh.NodeDef }

// PreRender is a core.Node method.
func (mesh *PolyMesh) PreRender() error {
	if err := mesh.init(); err != nil {
		return err
	}

	mesh.facecount = len(mesh.idxp) / 3
	mesh.vertidxstride = 3

	for _, shader := range mesh.Shader {
		if s := core.FindNode(shader); s != nil {
			shader, ok := s.(core.Shader)

			if !ok {
				return fmt.Errorf("Unable to find shader %v", shader)
			}

			mesh.shader = append(mesh.shader, shader)
		} else {
			return fmt.Errorf("Unable to find node (shader %v)", shader)

		}
	}

	return mesh.initAccel()
}

// PostRender is a core.Node method.
func (mesh *PolyMesh) PostRender() error { return nil }

// MotionKeys returns the number of motion keys.
func (mesh *PolyMesh) MotionKeys() int {
	//return len(mesh.Transform.Elems)

	if mesh.accel.qbvh != nil {
		return 1
	}

	return len(mesh.accel.mqbvh.Boxes)

}

func create() (core.Node, error) {
	mfile := PolyMesh{IsVisible: true}

	return &mfile, nil
}

func init() {
	nodes.Register("PolyMesh", create)
}
